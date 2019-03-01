package es

import (
	"errors"
	"fmt"
	"reflect"
)

// get payload struct type of  event
func eventPayloadType(event IESEvent) (reflect.Type, error) {

	eventType := reflect.TypeOf(event)

	// ensure property exists
	field, exists := eventType.FieldByNameFunc(func(s string) bool {
		return s == "payload"
	})

	// exit in case the payload prop does not exist
	if !exists {
		return nil, errors.New("payload property does not exist")
	}

	// ensure that property is a struct
	if field.Type.Kind() != reflect.Struct {
		return nil, errors.New("the payload property must be struct")
	}

	return field.Type, nil

}

// check if the IESEvent implementation supports the "New" function
func doesEventHasValidNewMethod(event IESEvent) error {

	eventType := reflect.TypeOf(event)

	// exit if doesn't implement at all
	method, exists := eventType.MethodByName("Factory")
	if !exists {
		return errors.New("event doesn't support 'New' method")
	}

	// the New method is supposed to expect two arguments
	if 2 != eventType.NumIn() {
		return errors.New("the 'New' method should expect to receive two arguments (first arg = instance of ESEvent | second arg = instance of the event payload")
	}

	// exit with invalid return arguments
	if 1 != eventType.NumOut() {
		return errors.New("the 'New' must return an instance of the IESEvent implementation")
	}

	// ensure that the first expected arg is of type ESEvent
	firstIn := method.Type.In(0)
	if firstIn != reflect.TypeOf(ESEvent{}) {
		return errors.New("the first expected parameter must be an non pointer ESEvent")
	}

	// payload
	payloadType, err := eventPayloadType(event)
	if err != nil {
		return err
	}

	// ensure that the second expected arg is of the same type as the payload property
	secondIn := method.Type.In(1)
	if secondIn != payloadType {
		return errors.New("the second expected parameter must be the same as the payload type")
	}

	// ensure that the struct returned from new is the same struct as the implementation of the passed event
	if method.Type.Out(0) != eventType {
		return errors.New("'Factory' method must return IESEvent implementation")
	}

	return nil

}

func reactorFactory(reactor interface{}) reactor {

	reactorType := reflect.TypeOf(reactor)

	handleMethod, exists := reactorType.MethodByName("Handle")
	if !exists {
		panic(fmt.Sprintf("reactor: %s doesn't have a 'Handle' method", reactorType.Name()))
	}

	return func(event IESEvent) {

		handleMethod.Func.Call([]reflect.Value{
			reflect.ValueOf(event),
		})

	}

}

func marshalEventPayload() (map[string]interface{}, error) {
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
