package reactor

import (
	"errors"
	"fmt"
	"github.com/florianlenz/event-sourcing-go/event"
	"reflect"
	"sync"
)

type reactor = func(event event.IESEvent)

type Registry struct {
	lock     *sync.Mutex
	reactors map[reflect.Type][]reflect.Value
}

// Register a new reactor
func (r *Registry) Register(reactor interface{}) error {

	// reactor type
	reactorValue := reflect.ValueOf(reactor)
	reactorType := reactorValue.Type()
	// @todo we have to pass in pointers in order to make it work. Google why we have too. I am not quite sure.
	if reactorType.Kind() != reflect.Ptr {
		return errors.New("invalid reactor - you have to pass in a pointer to the reactor")
	}
	reactorTypeElem := reactorType.Elem()

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
	for _, reactors := range r.reactors {
		for _, reactor := range reactors {
			if reactor.Type() == reactorType {
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
	handleMethodEvent := handleMethod.Type.In(1)
	if !handleMethodEvent.Implements(reflect.TypeOf((*event.IESEvent)(nil)).Elem()) {
		return fmt.Errorf("the handle method expects '%s' which is not an IESImplementation", handleMethodEvent.Name())
	}

	if handleMethodEvent.Kind() == reflect.Ptr {
		handleMethodEvent = handleMethodEvent.Elem()
	}

	// append reactor
	r.reactors[handleMethodEvent] = append(r.reactors[handleMethodEvent], reactorValue)

	return nil

}

// Fetch reactors for event
func (r *Registry) Reactors(e event.IESEvent) []reactor {

	// lock
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
	}()

	// event type
	eventType := reflect.TypeOf(e)
	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}

	// reactor type factory
	reactorTypeFactory := func(reactor reflect.Value) reactor {
		return func(event event.IESEvent) {
			reactor.MethodByName("Handle").Call([]reflect.Value{
				reflect.ValueOf(event),
			})
		}
	}

	// get reactors for event type
	reactors := []reactor{}
	for _, reactorValue := range r.reactors[eventType] {
		reactors = append(reactors, reactorTypeFactory(reactorValue))
	}

	return reactors

}

func NewReactorRegistry() *Registry {
	return &Registry{
		lock:     &sync.Mutex{},
		reactors: map[reflect.Type][]reflect.Value{},
	}
}
