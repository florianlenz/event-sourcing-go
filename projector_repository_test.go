package es

import (
	"context"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestProjectorRepository(t *testing.T) {

	Convey("test projector repository", t, func() {

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

		Convey("update last handled event", func() {

			Convey("update should create record if it was never persisted before", func() {

				// database
				db, err := createDB()
				So(err, ShouldBeNil)

				// collections
				eventCollection := db.Collection("events")
				projectorCollection := db.Collection("projectors")

				// make sure that the projectors collection is empty
				count, err := db.Collection("projectors").Count(context.Background(), nil)
				So(count, ShouldEqual, 0)

				// projector repository
				projectorRepository := projectorRepository{
					projectorCollection: projectorCollection,
					eventCollection:     eventCollection,
				}

				// test event id
				eventID := primitive.NewObjectID()

				// update last handled event
				err = projectorRepository.UpdateLastHandledEvent(
					&testProjector{
						name: "com.projector",
					},
					event{
						ID: &eventID,
					},
				)
				So(err, ShouldBeNil)

				// fetch projector manually
				result := projectorCollection.FindOne(context.Background(), bson.M{
					"projector_name": "com.projector",
				})

				// decode projector
				fetchedProjector := &projector{}
				So(result.Decode(&fetchedProjector), ShouldBeNil)

				// make sure that projector is correct
				So(*fetchedProjector.LastProcessedEvent, ShouldEqual, eventID)
				So(fetchedProjector.Name, ShouldEqual, "com.projector")

			})

			Convey("update should update last processed event id", func() {

				// create db
				db, err := createDB()
				So(err, ShouldBeNil)

				// collections
				eventCollection := db.Collection("events")
				projectorCollection := db.Collection("projectors")

				eventID := primitive.NewObjectID()

				// insert test projector
				_, err = projectorCollection.InsertOne(context.Background(), bson.M{
					"projector_name":       "com.projector",
					"last_processed_event": eventID,
				})
				So(err, ShouldBeNil)

				// repository
				projectorRepository := &projectorRepository{
					eventCollection:     eventCollection,
					projectorCollection: projectorCollection,
				}

				// update projector repository
				newEventID := primitive.NewObjectID()
				err = projectorRepository.UpdateLastHandledEvent(
					&testProjector{},
					event{
						ID: &newEventID,
					},
				)
				So(err, ShouldBeNil)

				// fetch updated projector
				result := projectorCollection.FindOne(context.Background(), bson.M{
					"projector_name": "com.projector",
				})

				fetchedProjector := &projector{}

				So(result.Decode(&fetchedProjector), ShouldBeNil)
				So(*fetchedProjector.LastProcessedEvent, ShouldEqual, eventID)

			})

			Convey("update should not create a second record", func() {

				// create db
				db, err := createDB()
				So(err, ShouldBeNil)

				// collections
				eventCollection := db.Collection("events")
				projectorCollection := db.Collection("projectors")

				// make sure that there are no projectors
				projectorCount, err := projectorCollection.Count(context.Background(), bson.M{})
				So(projectorCount, ShouldEqual, 0)
				So(err, ShouldBeNil)

				// projector repository
				projRepo := projectorRepository{
					eventCollection:     eventCollection,
					projectorCollection: projectorCollection,
				}

				// test projector
				testProj := &testProjector{
					name: "com.projector",
				}

				// update the first time
				firstUpdateEventID := primitive.NewObjectID()
				err = projRepo.UpdateLastHandledEvent(
					testProj,
					event{
						ID: &firstUpdateEventID,
					},
				)
				So(err, ShouldBeNil)

				// update the second time
				secondUpdateEventID := primitive.NewObjectID()
				err = projRepo.UpdateLastHandledEvent(
					testProj,
					event{
						ID: &secondUpdateEventID,
					},
				)
				So(err, ShouldBeNil)

				// make sure that there is only one projector
				projectorCount, err = projectorCollection.Count(context.Background(), bson.M{})
				So(projectorCount, ShouldEqual, 1)
				So(err, ShouldBeNil)

			})

		})

		Convey("out of sync by", func() {

			Convey("in the case the projector doesn't exist it should still return all relevant events", func() {

				// create db
				db, err := createDB()
				So(err, ShouldBeNil)

				// collections
				eventCollection := db.Collection("events")
				projectorCollection := db.Collection("projectors")

				// projector repository
				projectorRepo := &projectorRepository{
					eventCollection:     eventCollection,
					projectorCollection: projectorCollection,
				}

				outOfSyncBy, err := projectorRepo.OutOfSyncBy(&testProjector{
					// @todo stopped here
				})
				So(err, ShouldBeNil)

			})

			Convey("count unprocessed events", func() {

			})

		})

	})

}
