package event

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// register an new event with it's factory (factory = function that creates the event)
func (r *Registry) RegisterEvent(eventName string, event IESEvent) error {

	//  validate event's "New" method
	if err := doesEventHasValidFactoryMethod(event); err != nil {
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

func (r *Registry) GetEventName(event IESEvent) (string, error) {

	eventType := reflect.TypeOf(event)

	// lock
	r.lock.Lock()
	go func() {
		r.lock.Unlock()
	}()

	for eventName, registeredEvent := range r.registeredEvents {

		registeredEventType := reflect.TypeOf(registeredEvent)

		if registeredEventType == eventType {
			return eventName, nil
		}

	}

	return "", errors.New("couldn't find event name for event")

}

func (r *Registry) EventToESEvent(e Event) (IESEvent, error) {

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
type Registry struct {
	lock             *sync.Mutex
	registeredEvents map[string]IESEvent
}

func NewEventRegistry() *Registry {

	// event registry
	reg := &Registry{
		lock:             &sync.Mutex{},
		registeredEvents: map[string]IESEvent{},
	}

	return reg

}
