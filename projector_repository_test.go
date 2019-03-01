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
			client, err := mongo.Connect(context.TODO(), "mongodb://localhost:8034")
			if err != nil {
				return nil, err
			}

			// database
			db := client.Database("godb")
			err = db.Drop(context.Background())

			return db, err
		}

		var eventFactory = func(eventCollection *mongo.Collection, eventName string) primitive.ObjectID {
			id, err := eventCollection.InsertOne(context.Background(), bson.M{
				"name": eventName,
			})
			So(err, ShouldBeNil)
			objId, k := id.InsertedID.(primitive.ObjectID)
			So(k, ShouldBeTrue)
			return objId
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
					"name": "com.projector",
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

				// create tests events
				eventFactory(eventCollection, "user.created")
				eventFactory(eventCollection, "user.updated")
				eventFactory(eventCollection, "user.updated")
				eventFactory(eventCollection, "user.updated")
				eventFactory(eventCollection, "user.deleted")
				eventFactory(eventCollection, "user.deleted")

				// projector repository
				projectorRepo := &projectorRepository{
					eventCollection:     eventCollection,
					projectorCollection: projectorCollection,
				}

				outOfSyncBy, err := projectorRepo.OutOfSyncBy(&testProjector{
					name:               "com.projector",
					interestedInEvents: []IESEvent{},
				})
				So(err, ShouldBeNil)
				So(outOfSyncBy, ShouldEqual, 4)

			})

			Convey("only unprocessed events", func() {

				// create db
				db, err := createDB()
				So(err, ShouldBeNil)

				// collections
				eventCollection := db.Collection("events")
				projectorCollection := db.Collection("projectors")

				// create tests events
				eventFactory(eventCollection, "user.created")
				lastIndexedEventID := eventFactory(eventCollection, "user.created")
				eventFactory(eventCollection, "user.updated")
				eventFactory(eventCollection, "user.updated")
				eventFactory(eventCollection, "user.updated")
				eventFactory(eventCollection, "user.created")
				eventFactory(eventCollection, "user.created")
				eventFactory(eventCollection, "user.deleted")

				// projector repository
				projectorRepo := &projectorRepository{
					eventCollection:     eventCollection,
					projectorCollection: projectorCollection,
				}

				// test projector
				proj := testProjector{
					name:               "com.projector",
					interestedInEvents: []IESEvent{},
				}

				// update last handled event
				err = projectorRepo.UpdateLastHandledEvent(&proj, event{
					ID: &lastIndexedEventID,
				})
				So(err, ShouldBeNil)

				// query for out of sync
				outOfSyncBy, err := projectorRepo.OutOfSyncBy(&proj)
				So(err, ShouldBeNil)
				// expect to be 5 since two of the 7 events we are interested in already got processed
				So(outOfSyncBy, ShouldEqual, 5)

			})

		})

		Convey("test new projector repository", func() {

			// create db
			db, err := createDB()
			So(err, ShouldBeNil)

			// collections
			eventCollection := db.Collection("events")
			projectorCollection := db.Collection("projectors")

			projectorRepo := newProjectorRepository(eventCollection, projectorCollection, NewEventRegistry())

			So(projectorRepo.eventCollection, ShouldEqual, eventCollection)
			So(projectorRepo.projectorCollection, ShouldEqual, projectorCollection)

		})

		Convey("drop all projectors", func() {

			// db
			db, err := createDB()
			So(err, ShouldBeNil)

			projectorCollection := db.Collection("projectors")

			// projector repository
			projectorRepository := newProjectorRepository(db.Collection("events"), projectorCollection, NewEventRegistry())

			// insert test projectors
			_, err = db.Collection("projectors").InsertMany(context.Background(), []interface{}{
				bson.M{},
				bson.M{},
				bson.M{},
			})
			So(err, ShouldBeNil)

			// make sure that the correct amount of projectors got inserted
			projectorCount, err := projectorCollection.Count(context.Background(), bson.M{})
			So(err, ShouldBeNil)
			So(projectorCount, ShouldEqual, 3)

			// drop whole collection
			So(projectorRepository.Drop(), ShouldBeNil)

			// ensure that there are no documents left in the collection
			projectorCount, err = projectorCollection.Count(context.Background(), bson.M{})
			So(err, ShouldBeNil)
			So(projectorCount, ShouldEqual, 0)

		})

	})

}
