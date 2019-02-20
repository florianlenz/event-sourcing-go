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
	// @todo maybe add check if projectors are empty in case this is true
	bypassOutOfSyncCheck bool) *Processor {

	stop := make(chan struct{})

	p := &Processor{
		stop:                stop,
		msgBus:              msgBus,
		projectorRegistry:   projectorRegistry,
		projectorRepository: projectorRepository,
	}

	waitForReady := make(chan struct{}, 1)

	go func() {

		occurredEventSubscriber := msgBus.Subscribe("event:occurred")

		// ready :)
		waitForReady <- struct{}{}

		for {

			select {

			// handle occurred event
			case value := <-occurredEventSubscriber:

				// @todo check if event already got handled

				// mark event as processed
				var processedEvent = func() {
					msgBus.Emit("event:processed", value)
				}

				// cast to event id
				eventIDStr, k := value.(string)
				if !k {
					logger.Error(fmt.Errorf("expected to received event ID, received: type: %T value: %v", value, value))
					processedEvent()
					continue
				}

				// decode object id
				eventID, err := primitive.ObjectIDFromHex(eventIDStr)
				if err != nil {
					logger.Error(fmt.Errorf("it seems like: %s is not a valid hex string. Original error: %s", eventIDStr, err.Error()))
					processedEvent()
					continue
				}

				// persisted event
				persistedEvent, err := eventRepository.FetchByID(eventID)
				if err != nil {
					logger.Error(err)
					processedEvent()
					continue
				}

				// transform persisted event to event sourcing event
				esEvent, err := eventRegistry.EventToESEvent(persistedEvent)
				if err != nil {
					logger.Error(err)
					processedEvent()
					continue
				}

				projectors := projectorRegistry.ProjectorsForEvent(esEvent)
				for _, projector := range projectors {

					if !bypassOutOfSyncCheck {

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

				processedEvent()

			// kill go routine
			case <-stop:
				msgBus.Unsubscribe(occurredEventSubscriber)
				return
			}

		}

	}()

	// wait till ready
	<-waitForReady

	return p

}
