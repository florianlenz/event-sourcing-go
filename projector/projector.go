package projector

import "es/event"

type IProjector interface {
	// unique name of the projector
	Name() string
	Handle(event event.IEvent) error
}
