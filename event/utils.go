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

func PayloadToMap(event IESEvent) (map[string]interface{}, error) {

	// event type
	eventType := reflect.TypeOf(event)
	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}

	// event value
	eventValue := reflect.ValueOf(event)
	if eventValue.Kind() == reflect.Ptr {
		eventValue = eventValue.Elem()
	}

	payloadValue := eventValue.FieldByName("Payload")
	if !eventValue.IsValid() {
		return nil, fmt.Errorf("payload field doesn't exist on event '%s'", eventType.Name())
	}

	// ensure field is not a pointer
	if eventValue.Type().Kind() == reflect.Ptr {
		return nil, fmt.Errorf("payload field of event '%s' is not supposed to be a pointer", eventType.Name())
	}

	// get payload type
	payloadType := payloadValue.Type()

	parameters := map[string]interface{}{}

	for i := 0; i < payloadType.NumField(); i++ {

		// value field
		field := payloadValue.Field(i)

		// type field
		typeField := payloadType.Field(i)

		// get name for payload
		fieldPayloadName, exists := typeField.Tag.Lookup("es")
		if !exists {
			return nil, fmt.Errorf("missing 'es' tag in events payload field (event: '%s', payload field: '%s')", eventType.Name(), typeField.Name)
		}

		switch typeField.Type.Kind() {

		// @ todo what about floats and the huge numbers? Need to validate how they are persisted with mongodb
		case reflect.String:
			parameters[fieldPayloadName] = field.Interface()
		case reflect.Bool:
			parameters[fieldPayloadName] = field.Interface()
		case reflect.Int:
			parameters[fieldPayloadName] = field.Interface()
		case reflect.Uint:
			parameters[fieldPayloadName] = field.Interface()
		case reflect.Float32:
			parameters[fieldPayloadName] = field.Interface()
		case reflect.Float64:
			parameters[fieldPayloadName] = field.Interface()
		case reflect.Struct:

			// cast to an implementation that can be marshaled
			// @todo it's not possible to cast to an instance of Marshal if receiver is a pointer
			m, k := field.Interface().(Marshal)
			if !k {
				return nil, fmt.Errorf("failed to marshal field: '%s' of payload: '%s' - doesn't statisfy Marshal interface", field.Type().Name(), payloadType.Name())
			}

			// marshal the field
			marshaledField, err := m.Marshal()
			if err != nil {
				return nil, fmt.Errorf("failed to marshal field: %s of payload: %s - original error: \"%s\"", field.Type().Name(), payloadType.Name(), err.Error())
			}

			parameters[fieldPayloadName] = marshaledField
		default:
			return nil, fmt.Errorf("type: %s is not supported (field: '%s' - payload: '%s')", typeField.Type.Kind(), field.Type().Name(), payloadType.Name())
		}

	}

	return parameters, nil
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

		// the field must be exported
		if !field.CanSet() {
			continue
		}

		// get tag
		fieldTag, exists := typeField.Tag.Lookup("es")
		if !exists {
			return reflect.Value{}, fmt.Errorf("missing 'es' tag in events payload field (event: '%s', payload field: '%s')", eventType.Name(), typeField.Name)
		}

		//
		payloadValue, exists := payload[fieldTag]
		if !exists {
			return reflect.Value{}, fmt.Errorf("failed to get value from payload for key: %s", fieldTag)
		}

		switch typeField.Type.Kind() {

		case reflect.String:

			// cast to string
			val, k := payloadValue.(string)
			if !k {
				return reflect.Value{}, fmt.Errorf("failed to cast field: '%s' of payload: '%s' to string", typeField.Name, payloadType.Name())
			}

			field.SetString(val)

		case reflect.Bool:

			// cast to bool
			val, k := payloadValue.(bool)
			if !k {
				return reflect.Value{}, fmt.Errorf("failed to cast field: '%s' of payload: '%s' to bool", typeField.Name, payloadType.Name())
			}

			field.SetBool(val)

		case reflect.Int:

			// cast to bool
			val, k := payloadValue.(int)
			if !k {
				return reflect.Value{}, fmt.Errorf("failed to cast field: '%s' of payload: '%s' to int", typeField.Name, payloadType.Name())
			}

			field.SetInt(int64(val))

		case reflect.Uint:

			// cast to bool
			val, k := payloadValue.(uint)
			if !k {
				return reflect.Value{}, fmt.Errorf("failed to cast field: '%s' of payload: '%s' to uint", typeField.Name, payloadType.Name())
			}

			field.SetUint(uint64(val))

		case reflect.Float32:

			// cast to float32
			val, k := payloadValue.(float64)
			if !k {
				return reflect.Value{}, fmt.Errorf("failed to cast field: '%s' of payload: '%s' to float32", typeField.Name, payloadType.Name())
			}

			field.SetFloat(val)

		case reflect.Float64:

			// cast to float64
			val, k := payloadValue.(float64)
			if !k {
				return reflect.Value{}, fmt.Errorf("failed to cast field: '%s' of payload: '%s' to float64", typeField.Name, payloadType.Name())
			}

			field.SetFloat(val)

		case reflect.Struct:

			val := reflect.New(field.Type())
			val = val.Elem()

			marshallable, k := val.Interface().(Marshal)
			if !k {
				return reflect.Value{}, fmt.Errorf("failed to unmarshal field: '%s' of payload: '%s' - doesn't statisfy Marshal interface", field.Type().Name(), payloadType.Name())
			}

			if err := marshallable.Unmarshal(payloadValue); err != nil {
				return reflect.Value{}, fmt.Errorf("failed to marshal field: %s of payload: %s - original error: \"%s\"", field.Type().Name(), payloadType.Name(), err.Error())
			}

			field.Set(val)

		default:
			return reflect.Value{}, fmt.Errorf("type: %s is not supported (field: '%s' - payload: '%s')", typeField.Type.Kind(), field.Type().Name(), payloadType.Name())
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
