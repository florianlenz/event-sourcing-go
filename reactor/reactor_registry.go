package reactor

import (
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
	reactorType := reflect.TypeOf(reactor)
	if reactorType.Kind() == reflect.Ptr {
		reactorType = reactorType.Elem()
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
				return fmt.Errorf("reactor '%s' has already been registered", reactorType.Name())
			}
		}
	}

	// create reactor
	cratedReactor, err := reactorFactory(reactor)
	if err != nil {
		return err
	}

	// ensure that reactor has a handle method
	method, exists := reactorType.MethodByName("Handle")
	if !exists {
		return fmt.Errorf("reactor '%s' must have a 'Handle' method", reactorType.Name())
	}

	firstParameterType := method.Type.In(0)

	// append reactor
	r.reactors[reactorType] = append(r.reactors[reactorType], registeredReactor{
		eventType: firstParameterType,
		reactor:   cratedReactor,
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
