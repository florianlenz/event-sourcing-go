package consistent

import (
	er "es/event_registry"
	mb "es/msgbus"
	pr "es/projector_registry"
	es "es/store"
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
				event, k := value.(es.Event)
				if !k {
					panic("expected to receive an event from the event store")
				}

				// transform persisted event to event sourcing event
				esEvent, err := eventRegistry.EventToESEvent(event)
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
					projectorRepository.UpdateLastHandledEvent(projector, event)

				}

				// emit event processed event
				msgBus.Emit("event:processed", event)

			// kill go routine
			case <-stop:
				msgBus.UnSubscribe(occurredEventSubscriber)
				break
			}

		}

	}()

	return p

}
