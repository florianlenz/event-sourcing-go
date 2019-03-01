package es

import (
	"errors"
	"reflect"
	"sync"
)

type reactor = func(event IESEvent)

type registeredReactor struct {
	eventType reflect.Type
	reactor   reactor
}

type ReactorRegistry struct {
	lock     *sync.Mutex
	reactors map[reflect.Type][]registeredReactor
}

// Register a new reactor
func (r *ReactorRegistry) Register(reactor interface{}) error {

	reactorType := reflect.TypeOf(reactor)

	// lock the application
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
	}()

	// ensure that reactor hasn't been added
	for _, registeredReactorCollection := range r.reactors {
		for _, registeredReactor := range registeredReactorCollection {
			if registeredReactor.eventType == reactorType {
				return errors.New("reactor has already been registered")
			}
		}
	}

	if reactorType.Kind() != reflect.Struct {
		return errors.New("reactor must be a struct")
	}

	method, exists := reactorType.MethodByName("Handle")
	if !exists {
		return errors.New("reactor must have a 'Handle' method")
	}

	if method.Type.NumIn() != 1 {
		return errors.New("the reactors 'Handle' method must accept one parameter")
	}

	firstParameterType := method.Type.In(0)

	if !firstParameterType.Implements(reflect.TypeOf(new(IESEvent))) {
		return errors.New("the reactors 'Handle' method must take an implementation of IESEvent as it's first parameter")
	}

	// create reactor
	cratedReactor, err := reactorFactory(reactor)
	if err != nil {
		return err
	}

	// append reactor
	r.reactors[reactorType] = append(r.reactors[reactorType], registeredReactor{
		eventType: firstParameterType,
		reactor:   cratedReactor,
	})

	return nil

}

// Fetch reactors for event
func (r *ReactorRegistry) Reactors(event IESEvent) []reactor {

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

func NewReactorRegistry() *ReactorRegistry {

	r := &ReactorRegistry{
		reactors: map[reflect.Type][]registeredReactor{},
	}

	return r

}
