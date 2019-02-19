package es

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

type event struct {
	ID         *primitive.ObjectID `bson:"_id,omitempty"`
	Name       string              `bson:"name"`
	Payload    Payload             `bson:"payload"`
	Version    uint8               `bson:"version"`
	OccurredAt int64               `bson:"occurred_at"`
}
