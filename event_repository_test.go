package es

import (
	"context"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestEventRepository(t *testing.T) {

	Convey("Event Repository", t, func() {

		var createDB = func() (*mongo.Database, error) {

			// create client
			client, err := mongo.Connect(context.TODO(), "mongodb://test11:test11@ds141815.mlab.com:41815/godb")
			if err != nil {
				return nil, err
			}

			// database
			db := client.Database("godb")
			err = db.Drop(context.Background())

			return db, err
		}

		Convey("save", func() {

			Convey("save successfully", func() {

				db, err := createDB()
				So(err, ShouldBeNil)

				eventRepository := eventRepository{
					db: db,
				}

				event := &event{
					Name: "user.created",
					Payload: Payload{
						"key": "value",
					},
					Version:    1,
					OccurredAt: time.Now().Unix(),
				}

				err = eventRepository.Save(event)
				So(err, ShouldBeNil)

			})

		})

		Convey("get by id", func() {

			Convey("try to fetch event that doesn't exist", func() {

				db, err := createDB()
				So(err, ShouldBeNil)

				eventID := primitive.NewObjectID()

				eventRepository := eventRepository{
					db: db,
				}

				fetchedEvent, err := eventRepository.FetchByID(eventID)
				So(err, ShouldBeError, "mongo: no documents in result")
				So(fetchedEvent, ShouldResemble, event{})

			})

			Convey("fetch successfully ", func() {

				db, err := createDB()
				So(err, ShouldBeNil)

				// event repo
				eventRepository := eventRepository{
					db: db,
				}

				// event
				e := &event{
					Name: "user.created",
					Payload: Payload{
						"key": "value",
					},
					Version:    1,
					OccurredAt: time.Now().Unix(),
				}

				// persist event
				So(eventRepository.Save(e), ShouldBeNil)

				// fetch the event by the id of the persisted event
				fetchedEvent, err := eventRepository.FetchByID(*e.ID)
				So(err, ShouldBeNil)

				// make sure the persisted and fetched event are the same
				So(*e, ShouldResemble, fetchedEvent)

			})

		})

	})

}
