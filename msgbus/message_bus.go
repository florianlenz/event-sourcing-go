package msgbus

type IMessageBus interface {
	// emit event (send it to the message bus)
	Emit(eventName string, event interface{})
	// subscribe to all events
	Subscribe(eventName string) <-chan interface{}
	// unsubscribe form event
	UnSubscribe(listener *chan interface{})
	// will return event listener that is de allocated once used
	Once(eventName string) chan interface{}
}
