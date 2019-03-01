package es

import (
	"errors"
	"fmt"
	"reflect"
)

// get payload struct type of  event
func eventPayloadType(event IESEvent) (reflect.Type, error) {

	eventType := reflect.TypeOf(event)

	// get element in case we received a pointer
	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}

	// ensure property exists
	field, exists := eventType.FieldByNameFunc(func(s string) bool {
		return s == "payload"
	})

	// exit in case the payload prop does not exist
	if !exists {
		return nil, fmt.Errorf("event: '%s' doesn't have a payload property - keep in mind that exported payload properties are not accepted", eventType.Name())
	}

	// ensure that property is a struct
	if field.Type.Kind() != reflect.Struct {
		return nil, fmt.Errorf("the payload of event '%s' must be a struct - got: '%s'", eventType.Name(), field.Type.Kind())
	}

	return field.Type, nil

}

// check if the IESEvent implementation supports the "Factory" function
func doesEventHasValidFactoryMethod(event IESEvent) error {

	eventType := reflect.TypeOf(event)

	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}

	// exit if doesn't implement a factory method at all
	method, exists := eventType.MethodByName("Factory")
	if !exists {
		return fmt.Errorf("event '%s' doesn't support 'Factory' method", eventType.Name())
	}

	// the Factory method is supposed to expect two arguments
	if 2 != eventType.NumIn() {
		return fmt.Errorf("the 'Factory' method of event '%s' should expect two parameters (first param = instance of ESEvent | second param = instance of the event payload", eventType.Name())
	}

	// exit with invalid return arguments
	if 1 != eventType.NumOut() {
		return fmt.Errorf("the 'Factory' method of event: '%s' must return an instance of a struct that implements IESEvent interface", eventType.Name())
	}

	// ensure that the first expected arg is of type ESEvent
	firstIn := method.Type.In(0)
	if firstIn != reflect.TypeOf(ESEvent{}) {
		return fmt.Errorf("the first expected parameter of the %s Factory method should be ESEvent - got %s", eventType.Name(), firstIn.Name())
	}

	// payload
	payloadType, err := eventPayloadType(event)
	if err != nil {
		return err
	}

	// ensure that the second expected arg is of the same type as the payload property
	secondIn := method.Type.In(1)
	if secondIn != payloadType {
		return fmt.Errorf("the second expected paramter type in the '%s' Factory method should be %s - got %s", eventType.Name(), payloadType.Name(), secondIn.Name())
	}

	// ensure that the struct returned from new is the same struct as the implementation of the passed event
	if method.Type.Out(0) != eventType {
		return fmt.Errorf("the Factory method of '%s' must return an instance of '%s'", eventType.Name(), eventType.Name())
	}

	return nil

}

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
	if !handleMethodParameterType.Implements(reflect.TypeOf(*new(IESEvent))) {
		return nil, fmt.Errorf("the handle method expects %s which is not an IESImplementation", handleMethodParameterType.Name())
	}

	return func(event IESEvent) {
		handleMethod.Func.Call([]reflect.Value{
			reflect.ValueOf(event),
		})
	}, nil

}

func marshalEventPayload(event IESEvent) (map[string]interface{}, error) {

	// get payload type
	payloadType, err := eventPayloadType(event)
	if err != nil {
		return nil, err
	}

	v := reflect.New(payloadType)

	parameters := map[string]interface{}{}

	for i := 0; i < v.NumField(); i++ {

		// value field
		field := v.Field(i)

		// type field
		typeField := payloadType.Field(i)

		// get name for payload
		fieldPayloadName, exists := typeField.Tag.Lookup("es")
		if !exists {
			return nil, err
		}

		switch typeField.Type.Kind() {

		// @ todo what about floats? and the huge numbers? Need to validate how they are persisted with mongodb
		case reflect.String:
		case reflect.Bool:
		case reflect.Int:
		case reflect.Int8:
		case reflect.Int16:
		case reflect.Int32:
		case reflect.Int64:
		case reflect.Uint:
		case reflect.Uint8:
		case reflect.Uint16:
		case reflect.Uint32:
		case reflect.Uint64:
		case reflect.Float32:
		case reflect.Float64:
		case reflect.Map:
			parameters[fieldPayloadName] = field.Interface()
		case reflect.Struct:

			// cast to an implementation that can be marshaled
			m, k := field.Interface().(Marshal)
			if !k {
				return nil, fmt.Errorf("failed to marshal field: %s of payload: %s", field.Type().Name(), payloadType.Name())
			}

			// marshal the field
			value, err := m.Marshal()
			if err != nil {
				return nil, fmt.Errorf("failed to marshal field: %s of payload: %s - original error: \"%s\"", field.Type().Name(), payloadType.Name(), err.Error())
			}

			parameters[fieldPayloadName] = value

		default:
			return nil, fmt.Errorf("type: %s is not supported (field: %s - payload: %s", typeField.Type.Kind(), field.Type().Name(), payloadType.Name())
		}

	}

	return map[string]interface{}{}, nil
}

// create payload type from payload
func payloadMapToPayload(payloadType reflect.Type, payload map[string]interface{}) (reflect.Value, error) {

	v := reflect.New(payloadType)

	for i := 0; i < v.NumField(); i++ {

		// value field
		field := v.Field(i)

		// type field
		typeField := payloadType.Field(i)

		//
		fieldTag, exists := typeField.Tag.Lookup("es")
		if !exists {
			return reflect.Value{}, fmt.Errorf("failed to lock up tag 'es' in field")
		}

		//
		payloadValue, exists := payload[fieldTag]
		if !exists {
			return reflect.Value{}, fmt.Errorf("failed to get value from payload for key: %s", fieldTag)
		}

		switch typeField.Type.Kind() {

		case reflect.String:

			// cast to string
			str, k := payloadValue.(string)
			if !k {
				return reflect.Value{}, errors.New("failed to ")
			}

			field.SetString(str)

		case reflect.Struct:

			// cast field ot Marshal interface
			v, k := field.Interface().(Marshal)
			if !k {
				return reflect.Value{}, errors.New("received struct in payload that doesn't implement the marshal interface")
			}

			// cast to param map
			params, k := payloadValue.(map[string]interface{})
			if !k {
				return reflect.Value{}, errors.New("failed to cast parameters ")
			}

			if err := v.Unmarshal(params); err != nil {
				return reflect.Value{}, err
			}

			// @todo unmarshal object
		default:
			return reflect.Value{}, errors.New("type is not supported")
		}

	}

	return v, nil

}

func callEventFactoryMethod(esEvent IESEvent, persistedEvent event) (IESEvent, error) {

	eventType := reflect.TypeOf(esEvent)

	// fetch method
	method, exists := eventType.MethodByName("Factory")
	if !exists {
		return nil, errors.New("the event implementation is missing the factory function - please add exported Factory function to your event")
	}

	// get payload type
	payloadType, err := eventPayloadType(esEvent)
	if err != nil {
		return nil, err
	}

	// compose Factory arguments
	funcArgs := []reflect.Value{
		reflect.ValueOf(ESEvent{
			occurredAt: persistedEvent.OccurredAt,
		}),
		reflect.ValueOf(payloadType),
	}

	out := method.Func.Call(funcArgs)

	createdEventValue := out[0]

	createdEvent, k := createdEventValue.Interface().(IESEvent)
	if !k {
		return nil, errors.New("created event doesn't match the expected event")
	}

	return createdEvent, nil

}
