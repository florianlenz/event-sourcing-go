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

// test event repository
type testEventRepository struct {
	save      func(event *event.Event) error
	fetchByID func(id primitive.ObjectID) (event.Event, error)
	cb        func(eventID primitive.ObjectID)
}

func (r testEventRepository) Map(cb func(eventID primitive.ObjectID)) error {
	r.cb = cb
	return nil
}

func (r *testEventRepository) Save(event *event.Event) error {
	return r.save(event)
}

func (r *testEventRepository) FetchByID(id primitive.ObjectID) (event.Event, error) {
	return r.fetchByID(id)
}

type testEvent struct {
	event.ESEvent
}

// test projector
type testProjector struct {
	name               string
	interestedInEvents []event.IESEvent
	handleEvent        func(event event.IESEvent) error
}

func (tp *testProjector) Name() string {
	return tp.name
}

func (tp *testProjector) InterestedInEvents() []event.IESEvent {
	return tp.interestedInEvents
}

func (tp *testProjector) Handle(event event.IESEvent) error {
	return tp.handleEvent(event)
}

// test logger
type testLogger struct {
	errorChan chan error
}

func (l *testLogger) Error(error error) {
	l.errorChan <- error
}

// @todo add test to make sure that out of sync check is enabled
func TestEventSourcing(t *testing.T) {

	var createDB = func() (*mongo.Database, error) {

		// create client
		client, err := mongo.Connect(context.TODO(), "mongodb://localhost:8034")
		if err != nil {
			return nil, err
		}

		// database
		db := client.Database("godb")
		err = db.Drop(context.Background())

		return db, err
	}

	Convey("event sourcing", t, func() {

		Convey("make sure that event is persisted", func() {

			db, err := createDB()
			So(err, ShouldBeNil)

			projectorRegistry := projector.NewProjectorRegistry()
			eventRegistry := event.NewEventRegistry()

			// persist event channel
			persistedEventChan := make(chan *event.Event, 1)

			// create event sourcing
			es := NewEventSourcing(nil, db, projectorRegistry, eventRegistry, reactor.NewReactorRegistry())
			es.eventRepository = &testEventRepository{
				save: func(event *event.Event) error {
					persistedEventChan <- event
					return errors.New("i am a test error")
				},
			}

			// test event to persist
			testEvent := &testEvent{}
			testEvent.ESEvent = event.NewESEvent(3333333, 2)

			// commit event
			So(es.Commit(testEvent, nil), ShouldBeError, "i am a test error")

			// wait till it reached the repository
			persistedEvent := <-persistedEventChan

			// ensure that the data is correct
			// @todo So(persistedEvent.Name, ShouldEqual, testEvent.name)
			//  @todo So(persistedEvent.Payload, ShouldResemble, testEvent.payload)
			So(persistedEvent.Version, ShouldResemble, testEvent.Version())
			So(persistedEvent.OccurredAt, ShouldResemble, time.Now().Unix())

		})

		Convey("ensure that projector onProcessed gets notified", func() {

			// db
			db, err := createDB()
			So(err, ShouldBeNil)

			// registries
			projectorRegistry := projector.NewProjectorRegistry()
			eventRegistry := event.NewEventRegistry()

			// register test projector
			err = projectorRegistry.Register(&testProjector{
				name: "",
			})

			// create event sourcing
			es := NewEventSourcing(&testLogger{errorChan: make(chan error, 10)}, db, projectorRegistry, eventRegistry, reactor.NewReactorRegistry())
			es.Start()

			// on processed channel
			onProcessed := make(chan struct{}, 1)

			// commit event
			So(es.Commit(testEvent{}, onProcessed), ShouldBeNil)

			// make sure that we waited till
			So(<-onProcessed, ShouldResemble, struct{}{})

		})

		Convey("make sure that onProcessed is only notified when the event actually got processed", func() {

		})

	})

}
