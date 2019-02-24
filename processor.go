package es

import (
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

type Processor struct {
	stop                chan struct{}
	projectorRegistry   *projectorRegistry
	projectorRepository iProjectorRepository
	eventQueue          chan processEvent
}

type processEvent struct {
	eventID     primitive.ObjectID
	onProcessed chan struct{}
}

func (p *Processor) Stop() {
	p.stop <- struct{}{}
}

func (p *Processor) Process(eventID primitive.ObjectID) <-chan struct{} {

	onProcessedChan := make(chan struct{}, 1)

	p.eventQueue <- processEvent{
		eventID:     eventID,
		onProcessed: onProcessedChan,
	}

	return onProcessedChan

}

func newProcessor(
	projectorRegistry *projectorRegistry,
	eventRegistry *EventRegistry,
	reactorRegistry *ReactorRegistry,
	projectorRepository iProjectorRepository,
	eventRepository iEventRepository,
	logger ILogger,
	replay bool) *Processor {

	stop := make(chan struct{})
	eventQueue := make(chan processEvent, 100)

	p := &Processor{
		stop:                stop,
		projectorRegistry:   projectorRegistry,
		projectorRepository: projectorRepository,
		eventQueue:          eventQueue,
	}

	go func() {

		for {

			select {

			// handle occurred event
			case processEvent := <-eventQueue:

				eventID := processEvent.eventID
				onProcessed := processEvent.onProcessed

				// @todo check if event already got handled

				// persisted event
				persistedEvent, err := eventRepository.FetchByID(eventID)
				if err != nil {
					logger.Error(err)
					onProcessed <- struct{}{}
					continue
				}

				// transform persisted event to event sourcing event
				esEvent, err := eventRegistry.EventToESEvent(persistedEvent)
				if err != nil {
					logger.Error(err)
					onProcessed <- struct{}{}
					continue
				}

				// project event
				projectors := projectorRegistry.ProjectorsForEvent(esEvent)
				for _, projector := range projectors {

					if !replay {

						// make sure that the projector is not out of sync
						outOfSyncBy, err := projectorRepository.OutOfSyncBy(projector)
						if err != nil {
							logger.Error(err)
							continue
						}

						// report if error is out of sync. Being out of sync by one is fine since we are about to process the event
						if outOfSyncBy > 1 {
							logger.Error(fmt.Errorf("projector '%s' is out of sync - tried to apply event with name '%s'", projector.Name(), esEvent.Name()))
							continue
						}

					}

					// handle event
					if err := projector.Handle(esEvent); err != nil {
						logger.Error(err)
						continue
					}

					// updated the last handled event on the projector
					err = projectorRepository.UpdateLastHandledEvent(projector, persistedEvent)
					if err != nil {
						logger.Error(err)
					}

				}

				// pass event to reactors
				if !replay {
					reactors := reactorRegistry.ForEvent(esEvent)

					for _, reactor := range reactors {
						reactor.Handle(esEvent)
					}

				}

				onProcessed <- struct{}{}

			// kill go routine
			case <-stop:
				return
			}

		}

	}()

	return p

}
