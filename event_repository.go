package es

import (
	"context"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type IEventRepository interface {
	// save event
	Save(event *event) error
	// fetch event by it's id
	FetchByID(id primitive.ObjectID) (event, error)
}

type EventRepository struct {
	db *mongo.Database
}

func (r *EventRepository) Save(event *event) error {
	events := r.db.Collection("events")
	_, err := events.InsertOne(context.Background(), event)
	return err
}

func (r *EventRepository) FetchByID(id primitive.ObjectID) (event, error) {

	// event collection
	events := r.db.Collection("events")

	// find event by it's id
	result := events.FindOne(context.Background(), bson.M{"_id": id})

	// decode event
	e := event{}
	if err := result.Decode(&e); err != nil {
		return event{}, err
	}

	return e, nil

}
