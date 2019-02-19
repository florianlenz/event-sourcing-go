package es

type IMessageBus interface {
	// emit event (send it to the message bus)
	Emit(eventName string, payload interface{})
	// subscribe to all events
	Subscribe(eventName string) Subscription
	// unsubscribe form event
	Unsubscribe(subscription Subscription)
}
