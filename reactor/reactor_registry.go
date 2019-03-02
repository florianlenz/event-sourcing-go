package reactor

import (
	"errors"
	"fmt"
	"github.com/florianlenz/event-sourcing-go/event"
	"reflect"
	"sync"
)

type reactor = func(event event.IESEvent)

type registeredReactor struct {
	eventType reflect.Type
	reactor   reactor
}

type Registry struct {
	lock     *sync.Mutex
	reactors map[reflect.Type][]registeredReactor
}

// Register a new reactor
func (r *Registry) Register(reactor interface{}) error {

	// reactor type
	reactorValue := reflect.ValueOf(reactor)
	reactorType := reactorValue.Type()
	reactorTypeElem := reactorType.Elem()
	// @todo we have to pass in pointers in order to make it work. Google why we have too. I am not quite sure.
	if reactorType.Kind() != reflect.Ptr {
		return errors.New("invalid reactor - you have to pass in a pointer to the reactor")
	}

	// exit if not a valid reactor
	if reactorTypeElem.Kind() != reflect.Struct {
		return fmt.Errorf("reactor '%s' must be a struct", reactorTypeElem.Name())
	}

	// lock the application
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
	}()

	// ensure that reactor hasn't been added
	for _, registeredReactorCollection := range r.reactors {
		for _, registeredReactor := range registeredReactorCollection {
			if registeredReactor.eventType == reactorType {
				return fmt.Errorf("reactor '%s' has already been registered", reactorTypeElem.Name())
			}
		}
	}

	// get handle method
	handleMethod, exists := reactorType.MethodByName("Handle")
	if !exists {
		return fmt.Errorf("reactor '%s' doesn't have a 'Handle' method", reactorTypeElem.Name())
	}

	// ensure that the handle method expects one argument
	// @todo figure out why this is two - makes no sense except for if the receiver is counted as an parameter too
	if handleMethod.Type.NumIn() != 2 {
		return fmt.Errorf("the handle method of reactor %s must expect exactly one parameter", reactorTypeElem.Name())
	}

	// ensure that the expected argument is an implementation of IESEvent
	handleMethodParameterType := handleMethod.Type.In(1)
	if !handleMethodParameterType.Implements(reflect.TypeOf((*event.IESEvent)(nil)).Elem()) {
		return fmt.Errorf("the handle method expects '%s' which is not an IESImplementation", handleMethodParameterType.Name())
	}

	firstParameterType := handleMethod.Type.In(0)

	// append reactor
	r.reactors[reactorType] = append(r.reactors[reactorType], registeredReactor{
		eventType: firstParameterType,
		reactor: func(event event.IESEvent) {
			reactorValue.MethodByName("Handle").Call([]reflect.Value{
				reflect.ValueOf(event),
			})
		},
	})

	return nil

}

// Fetch reactors for event
func (r *Registry) Reactors(event event.IESEvent) []reactor {

	// lock
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
	}()

	// event type
	eventType := reflect.TypeOf(event)

	// get reactors for event type
	reactors := []reactor{}
	registeredReactors := r.reactors[eventType]
	for _, registeredReactor := range registeredReactors {
		reactors = append(reactors, registeredReactor.reactor)
	}

	return reactors

}

func NewReactorRegistry() *Registry {
	return &Registry{
		lock:     &sync.Mutex{},
		reactors: map[reflect.Type][]registeredReactor{},
	}
}
