package projector

import "es/event"

type IProjector interface {
	// unique name of the projector
	Name() string
	// return the events the projector is interested in
	InterestedInEvents() []event.IESEvent
	// check if projector is interested in event
	InterestedInEvent(event event.IESEvent) bool
	// handle a given event sourcing event
	Handle(event event.IESEvent) error
}
