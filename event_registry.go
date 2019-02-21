package es

import (
	"fmt"
)

// will create an event from the payload. A factory is usually registered in the context of an event
type Factory func(payload Payload) IESEvent

// registered event
type registeredEvent struct {
	event   IESEvent
	factory Factory
}

type eventToESEvent struct {
	event    event
	response chan struct {
		esEvent IESEvent
		error   error
	}
}

// add event command for state container
type addEvent struct {
	event        IESEvent
	eventFactory Factory
	response     chan error
}

func (r *eventRegistry) RegisterEvent(event IESEvent, eventFactory Factory) error {

	responseChan := make(chan error, 1)

	r.addEvent <- addEvent{
		event:        event,
		eventFactory: eventFactory,
		response:     responseChan,
	}

	return <-responseChan

}

func (r *eventRegistry) EventToESEvent(e event) (IESEvent, error) {

	// response channel
	responseChan := make(chan struct {
		esEvent IESEvent
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
type eventRegistry struct {
	addEvent       chan addEvent
	eventToESEvent chan eventToESEvent
}

func newEventRegistry() *eventRegistry {

	// register event over this channel
	addEvent := make(chan addEvent)

	eventToESEvent := make(chan eventToESEvent)

	// event registry
	reg := &eventRegistry{
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
				if registered {
					addEvent.response <- fmt.Errorf("event with name '%s' got already registered", addEvent.event.Name())
					continue
				}

				// register event
				registeredEvents[addEvent.event.Name()] = registeredEvent{
					event:   addEvent.event,
					factory: addEvent.eventFactory,
				}

				addEvent.response <- nil

			case eventToESEvent := <-eventToESEvent:

				e := eventToESEvent.event

				// fetch registered event
				registeredEvent, exists := registeredEvents[e.Name]
				if !exists {
					eventToESEvent.response <- struct {
						esEvent IESEvent
						error   error
					}{esEvent: nil, error: fmt.Errorf("event with name '%s' hasn't been registered", e.Name)}
					continue
				}

				// create event from payload
				esEvent := registeredEvent.factory(e.Payload)

				// exit if invalid event is returned
				if esEvent == nil {
					eventToESEvent.response <- struct {
						esEvent IESEvent
						error   error
					}{esEvent: nil, error: fmt.Errorf("received nil from event factory - for event: %s", e.Name)}
					continue
				}

				// this is a case that shouldn't happen
				if esEvent.Name() != e.Name {
					eventToESEvent.response <- struct {
						esEvent IESEvent
						error   error
					}{esEvent: nil, error: fmt.Errorf("attention! the creation of an event with name '%s' resulted in the creation of an event with name: '%s'", e.Name, esEvent.Name())}
					continue
				}

				// send response back
				eventToESEvent.response <- struct {
					esEvent IESEvent
					error   error
				}{esEvent: esEvent, error: nil}

			}

		}

	}()

	return reg

}
