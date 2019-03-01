package es

import (
	"context"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
)

type iProjectorRepository interface {
	// check if projector is out of sync
	OutOfSyncBy(projector IProjector) (int64, error)
	// update the last handled event on the projector
	UpdateLastHandledEvent(projector IProjector, event event) error
	// drop projector collection
	Drop() error
}

type projectorRepository struct {
	eventCollection     *mongo.Collection
	projectorCollection *mongo.Collection
	eventRegistry       *EventRegistry
}

func (r *projectorRepository) UpdateLastHandledEvent(projector IProjector, event event) error {

	projectors := r.projectorCollection

	// create projector if it doesn't exist
	updateOptions := options.Update()
	updateOptions.SetUpsert(true)

	_, err := projectors.UpdateOne(
		context.Background(),
		bson.M{"name": projector.Name()},
		bson.M{
			"$set": bson.M{
				"last_processed_event": event.ID,
			},
		},
		updateOptions,
	)

	return err

}

func (r *projectorRepository) Drop() error {
	return r.projectorCollection.Drop(context.Background())
}

func (r *projectorRepository) OutOfSyncBy(p IProjector) (int64, error) {

	// event names that the projector subscribed to
	eventNames := bson.A{}
	for _, event := range p.InterestedInEvents() {
		eventName, err := r.eventRegistry.GetEventName(event)
		if err != nil {
			return 0, err
		}
		eventNames = append(eventNames, eventName)
	}

	// fetch projector
	result := r.projectorCollection.FindOne(context.Background(), bson.M{
		"name": p.Name(),
	})

	fetchedProjector := &projector{}

	// decode fetched projector
	err := result.Decode(fetchedProjector)
	switch err {
	case nil:
		outOfSyncBy, err := r.eventCollection.Count(context.Background(), bson.M{
			"name": bson.M{
				"$in": eventNames,
			},
			"_id": bson.M{
				"$gt": fetchedProjector.LastProcessedEvent,
			},
		})
		return outOfSyncBy, err
	case mongo.ErrNoDocuments:
		outOfSyncBy, err := r.eventCollection.Count(context.Background(), bson.M{
			"name": bson.M{
				"$in": eventNames,
			},
		})
		return outOfSyncBy, err
	default:
		return 0, err
	}

}

func newProjectorRepository(eventCollection, projectorCollection *mongo.Collection, eventRegistry *EventRegistry) *projectorRepository {
	return &projectorRepository{
		eventCollection:     eventCollection,
		projectorCollection: projectorCollection,
		eventRegistry:       eventRegistry,
	}
}
