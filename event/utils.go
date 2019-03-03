package event

import (
	"errors"
	"fmt"
	"reflect"
)

func validatePayloadField(event IESEvent) error {

	eventType := reflect.TypeOf(event)

	field, found := eventType.FieldByNameFunc(func(s string) bool {
		return s == "Payload"
	})

	if !found {
		return fmt.Errorf("failed to find Payload field in event: '%s'", eventType.Name())
	}

	if field.Type.Kind() == reflect.Ptr {
		return fmt.Errorf("payload of event: '%s' must not be a pointer", eventType.Name())
	}

	return nil

}

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

func MarshalEventPayload(event IESEvent) (map[string]interface{}, error) {

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
func payloadMapToPayload(event IESEvent, payload map[string]interface{}) (reflect.Value, error) {

	// validate event payload
	eventType := reflect.TypeOf(event)
	payloadField, exists := eventType.FieldByName("Payload")
	if !exists {
		panic("payload must exist")
	}

	payloadType := payloadField.Type

	newPayload := reflect.New(payloadType)
	if newPayload.Kind() == reflect.Ptr {
		newPayload = newPayload.Elem()
	}

	for i := 0; i < payloadType.NumField(); i++ {

		// value field
		field := newPayload.Field(i)

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

	return newPayload, nil

}

func createIESEvent(esEvent IESEvent, persistedEvent Event) (IESEvent, error) {

	// create new event instance
	newEvent := reflect.New(reflect.TypeOf(esEvent))
	if newEvent.Kind() == reflect.Ptr {
		newEvent = newEvent.Elem()
	}

	// set payload field
	payload, err := payloadMapToPayload(esEvent, persistedEvent.Payload)
	if err != nil {
		return nil, err
	}
	payloadField := newEvent.FieldByName("Payload")
	payloadField.Set(payload)

	// @todo add ESEvent creation

	// cast event to IESEvent
	createdEvent, k := newEvent.Interface().(IESEvent)
	if !k {
		return nil, errors.New("created event doesn't match the expected event")
	}

	return createdEvent, nil

}
