package es

import "es/event"

type testEvent struct {
	name       string
	payload    event.Payload
	version    uint8
	occurredAt int64
}

func (e testEvent) Name() string {
	return e.name
}

func (e testEvent) Payload() event.Payload {
	return e.payload
}

func (e testEvent) Version() uint8 {
	return e.version
}

func (e testEvent) OccurredAt() int64 {
	return e.occurredAt
}
