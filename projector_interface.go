package es

type IProjector interface {
	// unique name of the projector
	Name() string
	// return the events the projector is interested in
	InterestedInEvents() []IESEvent
	// handle a given event sourcing event
	Handle(event IESEvent) error
}
