package es

import (
	"fmt"
	"sync"
)

// register an new event with it's factory (factory = function that creates the event)
func (r *EventRegistry) RegisterEvent(eventName string, event IESEvent) error {

	//  validate event's "New" method
	if err := doesEventHasValidNewMethod(event); err != nil {
		return err
	}

	// lock / unlock
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
	}()

	// ensure event hasn't been added
	if _, registered := r.registeredEvents[eventName]; registered {
		return fmt.Errorf("%s has already been registered", eventName)
	}

	// register event
	r.registeredEvents[eventName] = event

	return nil

}

func (r *EventRegistry) EventToESEvent(e event) (IESEvent, error) {

	// lock / unlock
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
	}()

	// fetch registered event
	esEvent, exists := r.registeredEvents[e.Name]
	if !exists {
		return nil, fmt.Errorf("an event struct for event with name: %s hasn't been registered", esEvent)
	}

	// get the events payload type
	return callEventFactoryMethod(esEvent, e)

}

// the public registry it self
type EventRegistry struct {
	lock             *sync.Mutex
	registeredEvents map[string]IESEvent
}

func NewEventRegistry() *EventRegistry {

	// event registry
	reg := &EventRegistry{
		lock: &sync.Mutex{},
	}

	return reg

}
