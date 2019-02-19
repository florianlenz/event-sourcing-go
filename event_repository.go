package es

import (
	"context"
	"errors"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type iEventRepository interface {
	// save event
	Save(event *event) error
	// fetch event by it's id
	FetchByID(id primitive.ObjectID) (event, error)
}

type eventRepository struct {
	db *mongo.Database
}

func (r *eventRepository) Save(event *event) error {

	// events collection
	events := r.db.Collection("events")

	// insert the event
	insertionResult, err := events.InsertOne(context.Background(), event)
	if err != nil {
		return err
	}

	// cast to ObjectID
	id, k := insertionResult.InsertedID.(primitive.ObjectID)
	if !k {
		return errors.New("failed to persist event - no id in insertion response")
	}

	event.ID = &id
	return err
}

func (r *eventRepository) FetchByID(id primitive.ObjectID) (event, error) {

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

func newEventRepository(db *mongo.Database) *eventRepository {
	return &eventRepository{
		db: db,
	}
}
