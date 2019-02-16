package event

// event interface
type IESEvent interface {
	// name of the event
	Name() string
	// event payload in a map
	Payload() Payload
	// version of the event
	Version() uint8
	// commit date
	occurredAt() int64
}
