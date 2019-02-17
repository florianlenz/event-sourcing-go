package consistent

import (
	"es"
	er "es/event_registry"
	mb "es/msgbus"
	pr "es/projector_registry"
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
					panic("expected to receive an event")
				}

				// persisted event
				persistedEvent, err := eventRepository.FetchByID(eventID)
				if err != nil {
					panic(err)
				}

				// transform persisted event to event sourcing event
				esEvent, err := eventRegistry.EventToESEvent(*persistedEvent)
				if err != nil {
					// @todo log
					panic(err)
				}

				projectors := projectorRegistry.ProjectorsForEvent(esEvent)
				for _, projector := range projectors {

					// make sure that the projector is not out of sync
					if true == projectorRepository.OutOfSync(projector) && bypassOutOfSyncCheck == false {
						// @todo report that projector is out of sync
						continue
					}

					// handle event
					err = projector.Handle(esEvent)
					if err != nil {
						// @todo report error?
						panic(err)
					}

					// updated the last handled event on the projector
					projectorRepository.UpdateLastHandledEvent(projector, *persistedEvent)

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
