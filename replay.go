package es

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
)

func Replay(logger ILogger, db *mongo.Database, projectorRegistry *projectorRegistry, eventRegistry *eventRegistry) <-chan error {

	// collections
	eventCollection := db.Collection("events")
	projectorCollection := db.Collection("projectors")

	// repositories
	eventRepository := newEventRepository(eventCollection)
	projectorRepository := newProjectorRepository(eventCollection, projectorCollection)

	done := make(chan error, 1)

	// processor (nill is passed for the reactor registry since we don't need it when we replay)
	processor := newProcessor(projectorRegistry, eventRegistry, nil, projectorRepository, eventRepository, logger, true)

	// drop all projectors
	if err := projectorRepository.Drop(); err != nil {
		done <- err
		return done
	}

	// start background re playing
	go func() {

		// map over the events an project them
		err := eventRepository.Map(func(eventID primitive.ObjectID) {

			// tell processor to process event
			onProcessed := processor.Process(eventID)

			// wait till it got processed
			<-onProcessed

		})

		// finish replay
		done <- err

		// stop processor
		processor.Stop()

	}()

	return done

}
