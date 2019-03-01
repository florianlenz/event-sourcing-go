package reactor

import (
	"fmt"
	"github.com/florianlenz/event-sourcing-go/event"
	"reflect"
)

func reactorFactory(reactor interface{}) (reactor, error) {

	// get type of reactor
	reactorType := reflect.TypeOf(reactor)
	if reactorType.Kind() == reflect.Ptr {
		reactorType = reactorType.Elem()
	}

	// exit if not a valid reactor
	if reactorType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("reactor '%s' is not a struct", reactorType.Name())
	}

	// get handle method
	handleMethod, exists := reactorType.MethodByName("Handle")
	if !exists {
		return nil, fmt.Errorf("reactor '%s' doesn't have a 'Handle' method", reactorType.Name())
	}

	// ensure that the handle method expects one argument
	if handleMethod.Type.NumIn() != 1 {
		return nil, fmt.Errorf("the handle method of reactor %s must expect exactly one parameter", reactorType.Name())
	}

	// ensure that the expected argument is an implementation of IESEvent
	handleMethodParameterType := handleMethod.Type.In(0)
	if !handleMethodParameterType.Implements(reflect.TypeOf(*new(event.IESEvent))) {
		return nil, fmt.Errorf("the handle method expects %s which is not an IESImplementation", handleMethodParameterType.Name())
	}

	return func(event event.IESEvent) {
		handleMethod.Func.Call([]reflect.Value{
			reflect.ValueOf(event),
		})
	}, nil

}
