package event

import (
	"context"
	"errors"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type IEventRepository interface {
	// save event
	Save(event *Event) error
	// fetch event by it's id
	FetchByID(id primitive.ObjectID) (Event, error)
	// map over all events
	Map(cb func(eventID primitive.ObjectID)) error
}

type eventRepository struct {
	eventCollection *mongo.Collection
}

func (r *eventRepository) Save(event *Event) error {

	// insert the event
	insertionResult, err := r.eventCollection.InsertOne(context.Background(), event)
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

func (r *eventRepository) Map(cb func(primitive.ObjectID)) error {

	// create event cursor
	cursor, err := r.eventCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return err
	}

	ctx := context.Background()

	// start iterating over the event
	for cursor.Next(ctx) {
		event := &Event{}
		// @todo would be better if we would receive only the id
		if err := cursor.Decode(&event); err != nil {
			return err
		}
		cb(*event.ID)
	}

	return cursor.Close(ctx)

}

func (r *eventRepository) FetchByID(id primitive.ObjectID) (Event, error) {

	// find event by it's id
	result := r.eventCollection.FindOne(context.Background(), bson.M{"_id": id})

	// decode event
	e := Event{}
	if err := result.Decode(&e); err != nil {
		return Event{}, err
	}

	return e, nil

}

func NewEventRepository(eventCollection *mongo.Collection) *eventRepository {
	return &eventRepository{
		eventCollection: eventCollection,
	}
}
