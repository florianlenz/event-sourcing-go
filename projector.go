package es

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

type IProjectorRepository interface {
	// check if projector is out of sync
	OutOfSyncBy(projector Projector) (uint, error)
	// update the last handled event on the projector
	UpdateLastHandledEvent(projector *Projector, event event) error
	// fetch projector
	GetOrCreateProjector(projector IProjector) (*Projector, error)
	// persist projector
	Save(projector *Projector) error
}

// projector entry in the database
type Projector struct {
	ID                 primitive.ObjectID `bson:"_id"`
	ProjectorName      string             `bson:"projector_name"`
	LastProcessedEvent primitive.ObjectID `bson:"last_processed_event"`
}
