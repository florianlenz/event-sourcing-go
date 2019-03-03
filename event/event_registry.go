package event

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// the public registry it self
type Registry struct {
	lock             *sync.Mutex
	registeredEvents map[string]IESEvent
}

// register an new event with it's factory (factory = function that creates the event)
func (r *Registry) RegisterEvent(eventName string, event IESEvent) error {

	// ensure that the event is registered as non pointer
	if reflect.TypeOf(event).Kind() == reflect.Ptr {
		return errors.New("can't register pointer")
	}

	//  ensure that event has exported payload filed
	if err := validatePayloadField(event); err != nil {
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
	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}

	// lock
	r.lock.Lock()
	go func() {
		r.lock.Unlock()
	}()

	for eventName, registeredEvent := range r.registeredEvents {

		// event type
		registeredEventType := reflect.TypeOf(registeredEvent)

		if registeredEventType == eventType {
			return eventName, nil
		}

	}

	return "", errors.New("event hasn't been registered")

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
		return nil, fmt.Errorf("event '%s' hasn't been registered", e.Name)
	}

	// get the events payload type
	return createIESEvent(esEvent, e)

}

func NewEventRegistry() *Registry {

	// event registry
	reg := &Registry{
		lock:             &sync.Mutex{},
		registeredEvents: map[string]IESEvent{},
	}

	return reg

}
