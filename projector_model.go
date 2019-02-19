package es

import "github.com/mongodb/mongo-go-driver/bson/primitive"

type projector struct {
	ID                 *primitive.ObjectID `bson:"_id,omitempty"`
	Name               string              `bson:"projector_name"`
	LastProcessedEvent *primitive.ObjectID `bson:"last_processed_event"`
}
