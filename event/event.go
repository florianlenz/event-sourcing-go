package event

// event interface
type IEvent interface {
	// name of the event
	Name() string
	// event payload in a map
	Payload() Payload
	// version of the event
	Version() uint8
	// commit date
	occurredAt() int64
}
