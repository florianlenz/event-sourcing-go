package msgbus

type IMessageBus interface {
	// emit event (send it to the message bus)
	Emit(eventName string, payload interface{})
	// subscribe to all events
	Subscribe(eventName string) <-chan interface{}
	// unsubscribe form event
	Unsubscribe(subscription <-chan interface{})
}
