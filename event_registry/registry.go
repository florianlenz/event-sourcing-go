package event_registry

import (
	"errors"
	"es/event"
	"es/store"
	"fmt"
)

// will create an event from the payload. A factory is usually registered in the context of an event
type EventFactory func(payload event.Payload) event.IESEvent

// registered event
type registeredEvent struct {
	event   event.IESEvent
	factory EventFactory
}

type eventToESEvent struct {
	event    store.Event
	response chan struct {
		esEvent event.IESEvent
		error   error
	}
}

// add event command for state container
type addEvent struct {
	event        event.IESEvent
	eventFactory EventFactory
	response     chan error
}

func (r *Registry) RegisterEvent(event event.IESEvent, eventFactory EventFactory) {

	r.addEvent <- addEvent{
		event:        event,
		eventFactory: eventFactory,
	}

}

func (r *Registry) EventToESEvent(e store.Event) (event.IESEvent, error) {

	// response channel
	responseChan := make(chan struct {
		esEvent event.IESEvent
		error   error
	}, 1)

	// ask state machine to convert the event to
	r.eventToESEvent <- eventToESEvent{
		event:    e,
		response: responseChan,
	}

	// response
	response := <-responseChan

	// return error if there is an error
	if response.error != nil {
		return nil, response.error
	}

	return response.esEvent, nil

}

// the public registry it self
type Registry struct {
	addEvent       chan addEvent
	eventToESEvent chan eventToESEvent
}

func New() *Registry {

	// register event over this channel
	addEvent := make(chan addEvent)

	eventToESEvent := make(chan eventToESEvent)

	// event registry
	reg := &Registry{
		addEvent:       addEvent,
		eventToESEvent: eventToESEvent,
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

			case eventToESEvent := <-eventToESEvent:

				e := eventToESEvent.event

				// fetch registered event
				registeredEvent, exists := registeredEvents[e.Name]
				if !exists {
					eventToESEvent.response <- struct {
						esEvent event.IESEvent
						error   error
					}{esEvent: nil, error: fmt.Errorf("no event registered for name: %s", e.Name)}
					continue
				}

				// create event from payload
				esEvent := registeredEvent.factory(e.Payload)

				// send response back
				eventToESEvent.response <- struct {
					esEvent event.IESEvent
					error   error
				}{esEvent: esEvent, error: nil}

			}

		}

	}()

	return reg

}
