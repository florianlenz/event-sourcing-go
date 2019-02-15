package event_registry

import "errors"

type addEvent struct {
	event        event.IEvent
	eventFactory event.FromPayloadFactory
	response     chan error
}

func New() *Registry {

	addEvent := make(chan addEvent)

	reg := &Registry{}

	go func() {

		registeredEvents := eventMap{}

		for {

			select {
			case addEvent := <-addEvent:
				// ensure that event hasn't been added
				if registeredEvents.AlreadyRegistered(addEvent.event) {
					addEvent.response <- errors.New("event already added")
					return
				}
				registeredEvents[addEvent.event.Name()] = addEvent.event

			}

		}

	}()

}

type Registry struct {
}

func (r *Registry) RegisterEvent(event event.IEvent) {

}
