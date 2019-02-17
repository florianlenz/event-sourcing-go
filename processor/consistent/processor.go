package consistent

import (
	"es"
	er "es/event_registry"
	mb "es/msgbus"
	pr "es/projector_registry"
	"fmt"
)

type Processor struct {
	msgBus              mb.IMessageBus
	stop                chan struct{}
	projectorRegistry   pr.Registry
	projectorRepository IProjectorRepository
}

func (p *Processor) Stop() {
	p.stop <- struct{}{}
}

func New(
	msgBus mb.IMessageBus,
	projectorRegistry pr.Registry,
	eventRegistry er.Registry,
	projectorRepository IProjectorRepository,
	eventRepository es.IEventRepository,
	logger es.ILogger,
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

				// cast to event
				eventID, k := value.(uint64)
				if !k {
					logger.Error(fmt.Errorf("expected to recevied event ID, received: '%v'", eventID))
					continue
				}

				// persisted event
				persistedEvent, err := eventRepository.FetchByID(eventID)
				if err != nil {
					logger.Error(err)
					continue
				}

				// transform persisted event to event sourcing event
				esEvent, err := eventRegistry.EventToESEvent(*persistedEvent)
				if err != nil {
					logger.Error(err)
					continue
				}

				projectors := projectorRegistry.ProjectorsForEvent(esEvent)
				for _, projector := range projectors {

					// fetch projector
					persistedProjector, err := projectorRepository.GetOrCreateProjector(projector)
					if err != nil {
						logger.Error(err)
						continue
					}

					// make sure that the projector is not out of sync
					outOfSyncBy, err := projectorRepository.OutOfSyncBy(*persistedProjector)
					if err != nil {
						logger.Error(err)
						continue
					}

					// report if error is out of sync. Being out of sync by one is fine since we are about to process the event
					if outOfSyncBy > 1 && bypassOutOfSyncCheck == false {
						logger.Error(fmt.Errorf("projector '%s' is out of sync - tried to apply event with id '%s' with type '%s'", projector.Name(), esEvent.Name()))
						continue
					}

					// handle event
					err = projector.Handle(esEvent)
					if err != nil {
						logger.Error(err)
						continue
					}

					// updated the last handled event on the projector
					err = projectorRepository.UpdateLastHandledEvent(persistedProjector, *persistedEvent)
					if err != nil {
						logger.Error(err)
					}

				}

				// emit event processed event
				msgBus.Emit("event:processed", eventID)

			// kill go routine
			case <-stop:
				msgBus.UnSubscribe(occurredEventSubscriber)
				break
			}

		}

	}()

	return p

}
