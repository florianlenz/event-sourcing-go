package es

import (
	"context"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

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

		// projector registry
		projectorRegistry := NewProjectorRegistry()

		//  register test event
		createdWithPayloadChan := make(chan map[string]interface{}, 2)
		eventRegistry := NewEventRegistry()
		err = eventRegistry.RegisterEvent(testEvent{name: "user.registered"}, func(payload Payload) IESEvent {
			createdWithPayloadChan <- payload
			return nil
		})
		So(err, ShouldBeNil)

		done := Replay(logger, db, projectorRegistry, eventRegistry)

		// wait till the replay is done
		<-done

		So(<-createdWithPayloadChan, ShouldResemble, map[string]interface{}{
			"event": "one",
		})
		So(<-createdWithPayloadChan, ShouldResemble, map[string]interface{}{
			"event": "two",
		})

	})

}
