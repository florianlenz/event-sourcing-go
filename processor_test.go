package es

import (
	"context"
	"errors"
	"github.com/florianlenz/event-sourcing-go/event"
	"github.com/florianlenz/event-sourcing-go/projector"
	"github.com/florianlenz/event-sourcing-go/reactor"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

// test projector repository
type testProjectorRepository struct {
	outOfSyncBy            func(projector projector.IProjector) (int64, error)
	updateLastHandledEvent func(projector projector.IProjector, event event.Event) error
	drop                   func() error
}

func (r *testProjectorRepository) OutOfSyncBy(projector projector.IProjector) (int64, error) {
	return r.outOfSyncBy(projector)
}

func (r *testProjectorRepository) UpdateLastHandledEvent(projector projector.IProjector, event event.Event) error {
	return r.updateLastHandledEvent(projector, event)
}

func (r *testProjectorRepository) Drop() error {
	return r.drop()
}

// test reactor
type testReactor struct {
	handle func(event event.IESEvent)
}

func (r *testReactor) Handle(event testEvent) {
	r.handle(event)
}

func TestProcessor(t *testing.T) {

	type processorTestSet struct {
		processor         *Processor
		logger            *testLogger
		projectorRegistry *projector.Registry
		eventRegistry     *event.Registry
		reactorRegistry   *reactor.Registry
	}

	var newProcessorTestSet = func(replay bool, eventRepository event.IEventRepository, projectorRepository projector.IProjectorRepository) (*processorTestSet, error) {

		// logger
		logger := &testLogger{
			errorChan: make(chan error, 10),
		}

		// create client
		client, err := mongo.Connect(context.TODO(), "mongodb://localhost:8034")
		if err != nil {
			return nil, err
		}

		// database
		db := client.Database("godb")
		err = db.Drop(context.Background())

		// projector registry
		projectorRegistry := projector.NewProjectorRegistry()

		// event registry
		eventRegistry := event.NewEventRegistry()

		//  reactor registry
		reactorRegistry := reactor.NewReactorRegistry()

		// projector repository
		if projectorRepository == nil {
			projectorRepository = projector.NewProjectorRepository(db.Collection("events"), db.Collection("projectors"), eventRegistry)
		}

		// create new event repository if no other got passed in
		if eventRepository == nil {
			eventRepository = event.NewEventRepository(db.Collection("events"))
		}

		processor := newProcessor(projectorRegistry, eventRegistry, reactorRegistry, projectorRepository, eventRepository, logger, replay)

		p := &processorTestSet{
			processor:         processor,
			logger:            logger,
			projectorRegistry: projectorRegistry,
			eventRegistry:     eventRegistry,
			reactorRegistry:   reactorRegistry,
		}

		return p, nil

	}

	Convey("test processor", t, func() {

		Convey("event must exist - the process should be aborted if it doesn't", func() {

			eventID := primitive.NewObjectID()

			repoCalledWith := make(chan primitive.ObjectID, 1)
			// test projector with special error
			eventRepository := &testEventRepository{
				fetchByID: func(id primitive.ObjectID) (event.Event, error) {
					repoCalledWith <- id
					return event.Event{}, errors.New("failed to fetch event by it's id")
				},
			}

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, eventRepository, nil)
			So(err, ShouldBeNil)
			processorTestSet.processor.Start()

			// local variables
			logger := processorTestSet.logger

			processorTestSet.processor.Process(eventID)

			// we expect an error since we instructed the mock to return an error
			So(<-logger.errorChan, ShouldBeError, "failed to fetch event by it's id")

			// make sure repository got called with the correct object id
			So(<-repoCalledWith, ShouldEqual, eventID)

		})

		Convey("event must be transformable to ESEvent", func() {

			eventID := primitive.NewObjectID()

			// event repo
			eventRepository := &testEventRepository{
				fetchByID: func(id primitive.ObjectID) (event.Event, error) {
					return event.Event{Name: "unregistered.event"}, nil
				},
			}

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, eventRepository, nil)
			So(err, ShouldBeNil)
			processorTestSet.processor.Start()

			// local variables
			logger := processorTestSet.logger

			// process event
			processorTestSet.processor.Process(eventID)

			// we expect an error since the returned event from the event repo was never registered
			So(<-logger.errorChan, ShouldBeError, "event 'unregistered.event' hasn't been registered")

		})

		Convey("the 'event:processed' event should be emitted even if there are no projectors to project it on", func() {

		})

		Convey("in the case the projector is out of sync by more than one an error should be logged", func() {

			eventID := primitive.NewObjectID()

			// test
			eventRepository := &testEventRepository{
				fetchByID: func(id primitive.ObjectID) (event.Event, error) {
					return event.Event{Name: "user.registered"}, nil
				},
			}

			// projector repository
			projectorRepository := &testProjectorRepository{
				outOfSyncBy: func(projector projector.IProjector) (i int64, e error) {
					return 2, nil
				},
			}

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, eventRepository, projectorRepository)
			So(err, ShouldBeNil)
			processorTestSet.processor.Start()

			// register event
			err = processorTestSet.eventRegistry.RegisterEvent("user.registered", testEvent{})
			So(err, ShouldBeNil)

			// register test projector
			calledHandleOnProjector := make(chan struct{}, 1)
			err = processorTestSet.projectorRegistry.Register(&testProjector{
				name: "user.projector",
				interestedInEvents: []event.IESEvent{
					&testEvent{},
				},
				handleEvent: func(event event.IESEvent) error {
					calledHandleOnProjector <- struct{}{}
					return nil
				},
			})
			So(err, ShouldBeNil)

			// local variables
			logger := processorTestSet.logger

			//
			onProcessed := processorTestSet.processor.Process(eventID)

			// make sure event got marked as processed
			So(<-onProcessed, ShouldResemble, struct{}{})

			// expect error since we return greater than one from the out of sync method
			So(<-logger.errorChan, ShouldBeError, "projector 'user.projector' is out of sync - tried to apply event with name 'user.registered'")

			// make sure event
			calledProjectorsHandleMethod := false
			select {
			case <-calledHandleOnProjector:
				calledProjectorsHandleMethod = true
			case <-time.After(time.Second):
			}
			So(calledProjectorsHandleMethod, ShouldBeFalse)

		})

		Convey("error during event handling should be logged and projector shouldn't be updated", func() {

			eventID := primitive.NewObjectID()

			// test event repository
			eventRepository := &testEventRepository{
				fetchByID: func(id primitive.ObjectID) (event.Event, error) {
					return event.Event{Name: "user.registered"}, nil
				},
			}

			// projector repository
			projectorRepository := &testProjectorRepository{
				outOfSyncBy: func(projector projector.IProjector) (i int64, e error) {
					return 1, nil
				},
				updateLastHandledEvent: func(projector projector.IProjector, event event.Event) error {
					panic("not supposed to call update last handled event")
					return nil
				},
			}

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, eventRepository, projectorRepository)
			So(err, ShouldBeNil)
			processorTestSet.processor.Start()

			// register event
			err = processorTestSet.eventRegistry.RegisterEvent("user.registered", testEvent{})
			So(err, ShouldBeNil)

			// register test projector
			err = processorTestSet.projectorRegistry.Register(&testProjector{
				name: "user.projector",
				interestedInEvents: []event.IESEvent{
					&testEvent{},
				},
				handleEvent: func(event event.IESEvent) error {
					return errors.New("error during handling event")
				},
			})
			So(err, ShouldBeNil)

			// local variables
			logger := processorTestSet.logger

			// process event
			onProcessed := processorTestSet.processor.Process(eventID)

			// make sure event got marked as processed
			So(<-onProcessed, ShouldResemble, struct{}{})

			// expect error returned from handler
			So(<-logger.errorChan, ShouldBeError, "error during handling event")

		})

		Convey("last handled event should be updated on projector if handler doesn't return an error", func() {

			eventID := primitive.NewObjectID()

			// test event repository
			eventRepository := &testEventRepository{
				fetchByID: func(id primitive.ObjectID) (event.Event, error) {
					return event.Event{Name: "user.registered"}, nil
				},
			}

			// projector repository
			updatedLastHandledEvent := make(chan struct{}, 1)
			projectorRepository := &testProjectorRepository{
				outOfSyncBy: func(projector projector.IProjector) (i int64, e error) {
					return 1, nil
				},
				updateLastHandledEvent: func(projector projector.IProjector, event event.Event) error {
					updatedLastHandledEvent <- struct{}{}
					return nil
				},
			}

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, eventRepository, projectorRepository)
			So(err, ShouldBeNil)
			processorTestSet.processor.Start()

			// register event
			err = processorTestSet.eventRegistry.RegisterEvent("user.registered", testEvent{})
			So(err, ShouldBeNil)

			// register test projector
			err = processorTestSet.projectorRegistry.Register(&testProjector{
				name: "user.projector",
				interestedInEvents: []event.IESEvent{
					&testEvent{},
				},
				handleEvent: func(event event.IESEvent) error {
					return nil
				},
			})
			So(err, ShouldBeNil)

			// process event
			onProcessed := processorTestSet.processor.Process(eventID)

			// make sure event got marked as processed
			So(<-onProcessed, ShouldResemble, struct{}{})

			// make sure update method got called
			So(<-updatedLastHandledEvent, ShouldResemble, struct{}{})

		})

		Convey("calling the stop function should shut down the processor", func() {

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, nil, nil)
			So(err, ShouldBeNil)

			// processor
			processor := processorTestSet.processor
			processor.Start()

			// emit event and wait till it got processed
			onProcessedFirstEvent := processor.Process(primitive.ObjectID{})
			So(<-onProcessedFirstEvent, ShouldResemble, struct{}{})

			// emit event second time to make sure that the process really works
			onProcessedSecondEvent := processor.Process(primitive.ObjectID{})
			So(<-onProcessedSecondEvent, ShouldResemble, struct{}{})

			// stop processor
			processorTestSet.processor.Stop()

			// emit event second time to make sure that the process really works
			onProcessedThirdEvent := processor.Process(primitive.ObjectID{})
			select {
			case <-onProcessedThirdEvent:
				panic("it seems like the processor is still running")
			case <-time.After(time.Second):

			}

		})

		Convey("should react on event", func() {

			// mock event repository
			eventRepo := &testEventRepository{
				fetchByID: func(id primitive.ObjectID) (event.Event, error) {
					return event.Event{
						Name: "user.created",
					}, nil
				},
			}

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, eventRepo, nil)
			So(err, ShouldBeNil)

			// processor
			processor := processorTestSet.processor
			processor.Start()

			// reactor registry
			reactorRegistry := processorTestSet.reactorRegistry

			// register event
			err = processorTestSet.eventRegistry.RegisterEvent("user.created", testEvent{})
			So(err, ShouldBeNil)

			// register event
			calledReactor := make(chan struct{}, 1)
			err = reactorRegistry.Register(&testReactor{
				handle: func(event event.IESEvent) {
					calledReactor <- struct{}{}
				},
			})
			So(err, ShouldBeNil)

			// process
			processor.Process(primitive.ObjectID{})

			// ensure that reactor got called
			So(<-calledReactor, ShouldResemble, struct{}{})

		})

		Convey("should not react on events when replaying", func() {

			// mock event repository
			eventRepo := &testEventRepository{
				fetchByID: func(id primitive.ObjectID) (event.Event, error) {
					return event.Event{
						Name: "user.created",
					}, nil
				},
			}

			// create new processor
			processorTestSet, err := newProcessorTestSet(true, eventRepo, nil)
			So(err, ShouldBeNil)

			// processor
			processor := processorTestSet.processor
			processor.Start()

			// reactor registry
			reactorRegistry := processorTestSet.reactorRegistry

			// register event
			err = processorTestSet.eventRegistry.RegisterEvent("user.created", testEvent{})
			So(err, ShouldBeNil)

			// register event
			calledReactor := make(chan struct{}, 1)
			err = reactorRegistry.Register(&testReactor{
				handle: func(event event.IESEvent) {
					calledReactor <- struct{}{}
				},
			})
			So(err, ShouldBeNil)

			// process and wait till done
			<-processor.Process(primitive.ObjectID{})

			// ensure that reactor got called
			select {
			case <-calledReactor:
				panic("called reactor")
			case <-time.After(time.Second * 2):

			}

		})

		Convey("start working only after the start signal", func() {

			// create new processor
			processorTestSet, err := newProcessorTestSet(true, nil, nil)
			So(err, ShouldBeNil)

			processor := processorTestSet.processor

			eventID := primitive.ObjectID{}

			onProcessed := processor.Process(eventID)

			select {
			case <-onProcessed:
				panic("didn't expect event to be processed since we didn't start the processor")
			case <-time.After(time.Second * 2):
			}

			processor.Start()

			So(<-onProcessed, ShouldResemble, struct{}{})

		})

	})

}
