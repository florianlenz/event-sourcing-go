package es

import (
	"context"
	"errors"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestProcessor(t *testing.T) {

	type processorTestSet struct {
		msgBus            IMessageBus
		processor         *Processor
		logger            *testLogger
		projectorRegistry *projectorRegistry
		eventRegistry     *eventRegistry
	}

	var newProcessorTestSet = func(byPassOutOfSyncCheck bool, eventRepository iEventRepository, projectorRepository iProjectorRepository) (*processorTestSet, error) {

		// logger
		logger := &testLogger{
			errorChan: make(chan error, 10),
		}

		// create client
		client, err := mongo.Connect(context.TODO(), "mongodb://test11:test11@ds141815.mlab.com:41815/godb")
		if err != nil {
			return nil, err
		}

		// database
		db := client.Database("godb")
		err = db.Drop(context.Background())

		// projector registry
		projectorRegistry := newProjectorRegistry()

		// event registry
		eventRegistry := newEventRegistry()

		// projector repository
		if projectorRepository == nil {
			projectorRepository = newProjectorRepository(db.Collection("events"), db.Collection("projectors"))
		}

		// create new event repository if no other got passed in
		if eventRepository == nil {
			eventRepository = newEventRepository(db)
		}

		// message bus
		mb := NewMemoryMessageBus()

		processor := NewSynchronousProcessor(mb, projectorRegistry, eventRegistry, projectorRepository, eventRepository, logger, byPassOutOfSyncCheck)

		p := &processorTestSet{
			processor:         processor,
			logger:            logger,
			projectorRegistry: projectorRegistry,
			msgBus:            mb,
			eventRegistry:     eventRegistry,
		}

		return p, nil

	}

	Convey("test processor", t, func() {

		Convey("event id must be a string", func() {

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, nil, nil)
			So(err, ShouldBeNil)

			// local variables
			mb := processorTestSet.msgBus
			logger := processorTestSet.logger

			// subscribed to event processed
			eventProcessedSubscr := mb.Subscribe("event:processed")

			// emit event with invalid payload
			mb.Emit("event:occurred", 2)

			// we expect an error since the payload of "event:occurred" must be a string
			So(<-logger.errorChan, ShouldBeError, "expected to received event ID, received: type: int value: 2")

			// make sure event got marked as processed
			So(<-eventProcessedSubscr, ShouldEqual, 2)

		})

		Convey("object id must be an object id as hex string", func() {

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, nil, nil)
			So(err, ShouldBeNil)

			// local variables
			mb := processorTestSet.msgBus
			logger := processorTestSet.logger

			// subscribed to event processed
			eventProcessedSubscr := mb.Subscribe("event:processed")

			// emit event with invalid payload
			mb.Emit("event:occurred", "i-am-not-a-hex-string")

			// we expect an error since the payload of "event:occurred" must also be a hex string
			So(<-logger.errorChan, ShouldBeError, "it seems like: i-am-not-a-hex-string is not a valid hex string. Original error: encoding/hex: invalid byte: U+0069 'i'")

			// make sure event got marked as processed
			So(<-eventProcessedSubscr, ShouldEqual, "i-am-not-a-hex-string")

		})

		Convey("event must exist - the process should be aborted if it doesn't", func() {

			eventID := primitive.NewObjectID()

			repoCalledWith := make(chan primitive.ObjectID, 1)
			// test projector with special error
			eventRepository := &testEventRepository{
				fetchByID: func(id primitive.ObjectID) (event, error) {
					repoCalledWith <- id
					return event{}, errors.New("failed to fetch event by it's id")
				},
			}

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, eventRepository, nil)
			So(err, ShouldBeNil)

			// local variables
			mb := processorTestSet.msgBus
			logger := processorTestSet.logger

			// subscribed to event processed
			eventProcessedSubscr := mb.Subscribe("event:processed")

			// emit event
			mb.Emit("event:occurred", eventID.Hex())

			// we expect an error since we instructed the mock to return an error
			So(<-logger.errorChan, ShouldBeError, "failed to fetch event by it's id")

			// make sure event got marked as processed
			So(<-eventProcessedSubscr, ShouldEqual, eventID.Hex())

			// make sure repository got called with the correct object id
			So(<-repoCalledWith, ShouldEqual, eventID)

		})

		Convey("event must be transformable to ESEvent", func() {

			eventID := primitive.NewObjectID()

			// event repo
			eventRepository := &testEventRepository{
				fetchByID: func(id primitive.ObjectID) (event, error) {
					return event{Name: "unregistered.event"}, nil
				},
			}

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, eventRepository, nil)
			So(err, ShouldBeNil)

			// local variables
			mb := processorTestSet.msgBus
			logger := processorTestSet.logger

			// subscribed to event processed
			eventProcessedSubscr := mb.Subscribe("event:processed")

			// emit event
			mb.Emit("event:occurred", eventID.Hex())

			// we expect an error since the returned event from the event repo was never registered
			So(<-logger.errorChan, ShouldBeError, "event with name 'unregistered.event' hasn't been registered")

			// make sure event got marked as processed
			So(<-eventProcessedSubscr, ShouldEqual, eventID.Hex())

		})

		Convey("the 'event:processed' event should be emitted even if there are no projectors to project it on", func() {

		})

		Convey("in the case the projector is out of sync by more than one an error should be logged", func() {

			eventID := primitive.NewObjectID()

			// test
			eventRepository := &testEventRepository{
				fetchByID: func(id primitive.ObjectID) (event, error) {
					return event{Name: "user.registered"}, nil
				},
			}

			// projector repository
			projectorRepository := &testProjectorRepository{
				outOfSyncBy: func(projector IProjector) (i int64, e error) {
					return 2, nil
				},
			}

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, eventRepository, projectorRepository)
			So(err, ShouldBeNil)

			// register event
			err = processorTestSet.eventRegistry.RegisterEvent(&testEvent{name: "user.registered"}, func(payload Payload) IESEvent {
				return &testEvent{name: "user.registered"}
			})
			So(err, ShouldBeNil)

			// register test projector
			calledHandleOnProjector := make(chan struct{}, 1)
			err = processorTestSet.projectorRegistry.Register(&testProjector{
				name: "user.projector",
				interestedInEvents: []IESEvent{
					&testEvent{
						name: "user.registered",
					},
				},
				handleEvent: func(event IESEvent) error {
					calledHandleOnProjector <- struct{}{}
					return nil
				},
			})
			So(err, ShouldBeNil)

			// local variables
			mb := processorTestSet.msgBus
			logger := processorTestSet.logger

			// subscribed to event processed
			eventProcessedSubscr := mb.Subscribe("event:processed")

			// emit event
			mb.Emit("event:occurred", eventID.Hex())

			// make sure event got marked as processed
			So(<-eventProcessedSubscr, ShouldEqual, eventID.Hex())

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
				fetchByID: func(id primitive.ObjectID) (event, error) {
					return event{Name: "user.registered"}, nil
				},
			}

			// projector repository
			projectorRepository := &testProjectorRepository{
				outOfSyncBy: func(projector IProjector) (i int64, e error) {
					return 1, nil
				},
				updateLastHandledEvent: func(projector IProjector, event event) error {
					panic("not supposed to call update last handled event")
					return nil
				},
			}

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, eventRepository, projectorRepository)
			So(err, ShouldBeNil)

			// register event
			err = processorTestSet.eventRegistry.RegisterEvent(&testEvent{name: "user.registered"}, func(payload Payload) IESEvent {
				return &testEvent{name: "user.registered"}
			})
			So(err, ShouldBeNil)

			// register test projector
			err = processorTestSet.projectorRegistry.Register(&testProjector{
				name: "user.projector",
				interestedInEvents: []IESEvent{
					&testEvent{
						name: "user.registered",
					},
				},
				handleEvent: func(event IESEvent) error {
					return errors.New("error during handling event")
				},
			})
			So(err, ShouldBeNil)

			// local variables
			mb := processorTestSet.msgBus
			logger := processorTestSet.logger

			// subscribed to event processed
			eventProcessedSubscr := mb.Subscribe("event:processed")

			// emit event
			mb.Emit("event:occurred", eventID.Hex())

			// make sure event got marked as processed
			So(<-eventProcessedSubscr, ShouldEqual, eventID.Hex())

			// expect error returned from handler
			So(<-logger.errorChan, ShouldBeError, "error during handling event")

		})

		Convey("last handled event should be updated on projector if handler doesn't return an error", func() {

			eventID := primitive.NewObjectID()

			// test event repository
			eventRepository := &testEventRepository{
				fetchByID: func(id primitive.ObjectID) (event, error) {
					return event{Name: "user.registered"}, nil
				},
			}

			// projector repository
			updatedLastHandledEvent := make(chan struct{}, 1)
			projectorRepository := &testProjectorRepository{
				outOfSyncBy: func(projector IProjector) (i int64, e error) {
					return 1, nil
				},
				updateLastHandledEvent: func(projector IProjector, event event) error {
					updatedLastHandledEvent <- struct{}{}
					return nil
				},
			}

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, eventRepository, projectorRepository)
			So(err, ShouldBeNil)

			// register event
			err = processorTestSet.eventRegistry.RegisterEvent(&testEvent{name: "user.registered"}, func(payload Payload) IESEvent {
				return &testEvent{name: "user.registered"}
			})
			So(err, ShouldBeNil)

			// register test projector
			err = processorTestSet.projectorRegistry.Register(&testProjector{
				name: "user.projector",
				interestedInEvents: []IESEvent{
					&testEvent{
						name: "user.registered",
					},
				},
				handleEvent: func(event IESEvent) error {
					return nil
				},
			})
			So(err, ShouldBeNil)

			// local variables
			mb := processorTestSet.msgBus

			// subscribed to event processed
			eventProcessedSubscr := mb.Subscribe("event:processed")

			// emit event
			mb.Emit("event:occurred", eventID.Hex())

			// make sure event got marked as processed
			So(<-eventProcessedSubscr, ShouldEqual, eventID.Hex())

			// make sure update method got called
			So(<-updatedLastHandledEvent, ShouldResemble, struct{}{})

		})

		Convey("calling the stop function should shut down the processor", func() {

			// create new processor
			processorTestSet, err := newProcessorTestSet(false, nil, nil)
			So(err, ShouldBeNil)

			// local variables
			mb := processorTestSet.msgBus

			// subscribed to event processed
			eventProcessedSubscr := mb.Subscribe("event:processed")

			// emit event and wait till it got processed
			mb.Emit("event:occurred", "event@1")
			So(<-eventProcessedSubscr, ShouldEqual, "event@1")

			// emit event second time to make sure that the process really works
			mb.Emit("event:occurred", "event@2")
			So(<-eventProcessedSubscr, ShouldEqual, "event@2")

			// stop processor
			processorTestSet.processor.Stop()

			// emit event second time to make sure that the process really works
			mb.Emit("event:occurred", "event@2")
			select {
			case <-eventProcessedSubscr:
				panic("it seems like the processor is still running")
			case <-time.After(time.Second):

			}

		})

	})

}
