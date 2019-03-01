package event

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

type Event struct {
	ID         *primitive.ObjectID    `bson:"_id,omitempty"`
	Name       string                 `bson:"name"`
	Payload    map[string]interface{} `bson:"payload"`
	Version    uint8                  `bson:"version"`
	OccurredAt int64                  `bson:"occurred_at"`
}
