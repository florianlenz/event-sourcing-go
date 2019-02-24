package es

import (
	"context"
	"errors"
	"github.com/mongodb/mongo-go-driver/mongo"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

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

			projectorRegistry := newProjectorRegistry()
			eventRegistry := NewEventRegistry()

			// persist event channel
			persistedEventChan := make(chan *event, 1)

			// create event sourcing
			es := NewEventSourcing(nil, db, projectorRegistry, eventRegistry, NewReactorRegistry())
			es.eventRepository = &testEventRepository{
				save: func(event *event) error {
					persistedEventChan <- event
					return errors.New("i am a test error")
				},
			}

			// test event to persist
			testEvent := &testEvent{
				name: "user.persisted",
				payload: map[string]interface{}{
					"key": "value",
				},
				version: 1,
			}

			// commit event
			So(es.Commit(testEvent, nil), ShouldBeError, "i am a test error")

			// wait till it reached the repository
			persistedEvent := <-persistedEventChan

			// ensure that the data is correct
			So(persistedEvent.Name, ShouldEqual, testEvent.name)
			So(persistedEvent.Payload, ShouldResemble, testEvent.payload)
			So(persistedEvent.Version, ShouldResemble, testEvent.version)
			So(persistedEvent.OccurredAt, ShouldResemble, time.Now().Unix())

		})

		Convey("ensure that projector onProcessed gets notified", func() {

			// db
			db, err := createDB()
			So(err, ShouldBeNil)

			// registries
			projectorRegistry := newProjectorRegistry()
			eventRegistry := NewEventRegistry()

			// register test projector
			err = projectorRegistry.Register(&testProjector{
				name: "",
			})

			// create event sourcing
			es := NewEventSourcing(&testLogger{errorChan: make(chan error, 10)}, db, projectorRegistry, eventRegistry, NewReactorRegistry())

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
