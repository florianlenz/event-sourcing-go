package es

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type ProjectorRepository struct {
	db *mongo.Database
}

func (r *ProjectorRepository) outOfSyncQuery(projector Projector) *gorm.DB {

	projectors := r.db.Collection("projectors")

}

func (r *ProjectorRepository) OutOfSyncBy(projector Projector) (uint, error) {
	var count uint
	err := r.outOfSyncQuery(projector).Count(&count).Error
	return count, err
}

func (r *ProjectorRepository) GetOrCreateProjector(projector IProjector) (Projector, error) {

}

func (r *ProjectorRepository) Save(projector Projector) error {

	projectors := r.db.Collection("projectors")

	_, err := projectors.InsertOne(context.Background(), bson.M{
		"projector_id":            ProjectorID,
		"last_processed_event_id": projector.LastProcessedEvent,
	})

	return err

}

func (r *ProjectorRepository) UpdateLastHandledEvent(projector *Projector, event es.Event) error {

	projectors := r.db.Collection("projectors")

	_, err := projectors.UpdateOne(
		context.Background(),
		bson.M{
			"projector_id": projector.ProjectorID,
		},
		bson.M{
			"last_processed_event": event.Payload,
		},
	)

	return err

}

func NewProjectorRepository(db *gorm.DB) *ProjectorRepository {

}
