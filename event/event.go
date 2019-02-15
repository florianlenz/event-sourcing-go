package event

import "errors"

type PersistedEvent struct {
	id         int
	payload    map[string]interface{}
	eventName  string
	occurredAt uint64
}

type FromPayloadFactory func(eventName string, payload SerializedEventPayload) IEvent

type Package struct {
	EventName          string
	FromPayloadFactory func(payload SerializedEventPayload) (IEvent, error)
}

type SerializedEventPayload map[string]interface{}

func (sep SerializedEventPayload) GetString(key string) (string, error) {

	email, k := sep[key].(string)
	if !k {
		return "", errors.New("expected e_mail_address to be of type string")
	}

	return email, nil

}

type IEvent interface {
	Name() string
	SerializePayload() SerializedEventPayload
	Version() uint8
}

type UserRegisteredEvent struct {
	eMailAddress string
	password     string
	username     string
}

func (e UserRegisteredEvent) Name() string {
	return "user.registered"
}

func (e UserRegisteredEvent) SerializePayload() SerializedEventPayload {
	return map[string]interface{}{
		"e_mail_address": e.eMailAddress,
		"password":       e.password,
		"username":       e.username,
	}
}

func (e UserRegisteredEvent) Version() uint8 {
	return 1
}

func NewUserRegisteredEvent(eMailAddress string, password string, username string) UserRegisteredEvent {

	return UserRegisteredEvent{
		eMailAddress: eMailAddress,
	}

}
