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
}

type projectorRepository struct {
	eventCollection     *mongo.Collection
	projectorCollection *mongo.Collection
}

func (r *projectorRepository) UpdateLastHandledEvent(projector IProjector, event event) error {

	projectors := r.projectorCollection

	// create projector if it doesn't exist
	updateOptions := options.Update()
	updateOptions.SetUpsert(true)

	_, err := projectors.UpdateOne(
		context.Background(),
		bson.M{"projector_name": projector.Name()},
		bson.M{
			"$set": bson.M{
				"last_processed_event": event.ID,
			},
		},
		updateOptions,
	)

	return err

}

func (r *projectorRepository) OutOfSyncBy(p IProjector) (int64, error) {

	// event names that the projector subscribed to
	eventNames := bson.A{}
	for _, event := range p.InterestedInEvents() {
		eventNames = append(eventNames, event.Name())
	}

	// fetch projector
	result := r.projectorCollection.FindOne(context.Background(), bson.M{
		"name": p.Name(),
	})

	fetchedProjector := &projector{}

	if err := result.Decode(fetchedProjector); err != nil {
		return 0, err
	}

	// event collection
	eventCollection := r.eventCollection

	outOfSyncBy, err := eventCollection.Count(context.Background(), bson.M{
		"name": bson.M{
			"$in": eventNames,
		},
		"_id": bson.M{
			"$gt": fetchedProjector.ID,
		},
	})

	return outOfSyncBy, err
}

func newProjectorRepository(eventCollection, projectorCollection *mongo.Collection) *projectorRepository {
	return &projectorRepository{
		eventCollection:     eventCollection,
		projectorCollection: projectorCollection,
	}
}
