package es

import (
	e "es/event"
)

type Event struct {
	ID         uint64
	Name       string
	Payload    e.Payload
	Version    uint8
	OccurredAt int64
}

type IEventRepository interface {
	// save event
	Save(event *Event) error
	// fetch event by it's id
	FetchByID(eventID uint64) (*Event, error)
}
