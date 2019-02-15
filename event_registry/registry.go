package event_registry

import (
	"errors"
	"es/event"
)

// will create an event from the payload. A factory is usually registered in the context of an event
type EventFactory func(payload event.Payload) event.IEvent

// registered event
type registeredEvent struct {
	event   event.IEvent
	factory EventFactory
}

// add event command for state container
type addEvent struct {
	event        event.IEvent
	eventFactory EventFactory
	response     chan error
}

// the public registry it self
type Registry struct {
	addEvent chan addEvent
}

func New() *Registry {

	// register event over this channel
	addEvent := make(chan addEvent)

	// event registry
	reg := &Registry{
		addEvent: addEvent,
	}

	go func() {

		registeredEvents := map[string]registeredEvent{}

		for {

			select {

			// add event to registry
			case addEvent := <-addEvent:

				// ensure that event hasn't been added
				_, registered := registeredEvents[addEvent.event.Name()]
				if !registered {
					addEvent.response <- errors.New("event already added")
					return
				}

				// register event
				registeredEvents[addEvent.event.Name()] = registeredEvent{
					event:   addEvent.event,
					factory: addEvent.eventFactory,
				}

			}

		}

	}()

	return reg

}

func (r *Registry) RegisterEvent(event event.IEvent, eventFactory EventFactory) {

	r.addEvent <- addEvent{
		event:        event,
		eventFactory: eventFactory,
	}

}
