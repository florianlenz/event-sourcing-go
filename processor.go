package es

import (
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

type Processor struct {
	msgBus              IMessageBus
	stop                chan struct{}
	projectorRegistry   *projectorRegistry
	projectorRepository iProjectorRepository
}

func (p *Processor) Stop() {
	p.stop <- struct{}{}
}

func NewSynchronousProcessor(
	msgBus IMessageBus,
	projectorRegistry *projectorRegistry,
	eventRegistry *eventRegistry,
	projectorRepository iProjectorRepository,
	eventRepository iEventRepository,
	logger ILogger,
	bypassOutOfSyncCheck bool) *Processor {

	stop := make(chan struct{})

	p := &Processor{
		stop:                stop,
		msgBus:              msgBus,
		projectorRegistry:   projectorRegistry,
		projectorRepository: projectorRepository,
	}

	go func() {

		occurredEventSubscriber := msgBus.Subscribe("event:occurred")

		for {

			select {

			// handle occurred event
			case value := <-occurredEventSubscriber:

				// cast to event id
				eventIDStr, k := value.(string)
				if !k {
					logger.Error(fmt.Errorf("expected to recevied event ID, received: '%v'", eventIDStr))
					continue
				}

				// decode object id
				eventID, err := primitive.ObjectIDFromHex(eventIDStr)
				if err != nil {
					logger.Error(err)
					continue
				}

				// persisted event
				persistedEvent, err := eventRepository.FetchByID(eventID)
				if err != nil {
					logger.Error(err)
					continue
				}

				// transform persisted event to event sourcing event
				esEvent, err := eventRegistry.EventToESEvent(persistedEvent)
				if err != nil {
					logger.Error(err)
					continue
				}

				projectors := projectorRegistry.ProjectorsForEvent(esEvent)
				for _, projector := range projectors {

					// make sure that the projector is not out of sync
					outOfSyncBy, err := projectorRepository.OutOfSyncBy(projector)
					if err != nil {
						logger.Error(err)
						continue
					}

					// report if error is out of sync. Being out of sync by one is fine since we are about to process the event
					if outOfSyncBy > 1 && bypassOutOfSyncCheck == false {
						logger.Error(fmt.Errorf("projector '%s' is out of sync - tried to apply event with name '%s'", projector.Name(), esEvent.Name()))
						continue
					}

					// handle event
					err = projector.Handle(esEvent)
					if err != nil {
						logger.Error(err)
						continue
					}

					// updated the last handled event on the projector
					err = projectorRepository.UpdateLastHandledEvent(projector, persistedEvent)
					if err != nil {
						logger.Error(err)
					}

				}

				// emit event processed event
				msgBus.Emit("event:processed", eventID)

			// kill go routine
			case <-stop:
				msgBus.Unsubscribe(occurredEventSubscriber)
				break
			}

		}

	}()

	return p

}
