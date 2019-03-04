package es

import (
	"context"
	"github.com/florianlenz/event-sourcing-go/event"
	"github.com/florianlenz/event-sourcing-go/projector"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type replayTestEventPayload struct {
	Event string `es:"event"`
}

type replayTestEvent struct {
	event.ESEvent
	Payload replayTestEventPayload
}

func TestReplay(t *testing.T) {

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

	Convey("must replay events", t, func() {

		// logger
		logger := &testLogger{
			errorChan: make(chan error, 10),
		}

		// db
		db, err := createDB()
		So(err, ShouldBeNil)
		eventCollection := db.Collection("events")

		// first event
		_, err = eventCollection.InsertOne(context.Background(), bson.M{
			"name": "user.registered",
			"payload": bson.M{
				"event": "one",
			},
		})
		So(err, ShouldBeNil)

		// second event
		_, err = eventCollection.InsertOne(context.Background(), bson.M{
			"name": "user.registered",
			"payload": bson.M{
				"event": "two",
			},
		})
		So(err, ShouldBeNil)

		// projected events channel
		projectedEvents := make(chan event.IESEvent, 2)

		// projector registry
		projectorRegistry := projector.NewProjectorRegistry()
		So(projectorRegistry.Register(&testProjector{
			name: "user_projector",
			interestedInEvents: []event.IESEvent{
				replayTestEvent{},
			},
			handleEvent: func(event event.IESEvent) error {
				projectedEvents <- event
				return nil
			},
		}), ShouldBeNil)

		//  register test event
		eventRegistry := event.NewEventRegistry()
		So(eventRegistry.RegisterEvent("user.registered", replayTestEvent{}), ShouldBeNil)

		done := Replay(logger, db, projectorRegistry, eventRegistry)

		// wait till the replay is done
		<-done

		So(<-projectedEvents, ShouldResemble, replayTestEvent{
			Payload: replayTestEventPayload{
				Event: "one",
			},
		})
		So(<-projectedEvents, ShouldResemble, replayTestEvent{
			Payload: replayTestEventPayload{
				Event: "two",
			},
		})

	})

}
