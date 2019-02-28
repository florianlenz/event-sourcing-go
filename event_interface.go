package es

// event interface
// ATTENTION: your event must include a "Factory" method which will receive an instance of ESEvent AND an instance of the event Payload. The New method must return a non pointer instance of "it self"
type IESEvent interface {
	// version of the event
	Version() uint8
	// commit date
	OccurredAt() int64
}
