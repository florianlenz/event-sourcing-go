package es

type emit struct {
	eventName string
	payload   interface{}
}

type subscribe struct {
	eventName string
	response  chan Subscription
}

type Subscription <-chan interface{}

type MemoryMessageBus struct {
	close              chan struct{}
	eventChannel       chan emit
	subscribeChannel   chan subscribe
	unsubscribeChannel chan Subscription
}

func (b *MemoryMessageBus) Emit(eventName string, payload interface{}) {
	b.eventChannel <- emit{
		eventName: eventName,
		payload:   payload,
	}
}

func (b *MemoryMessageBus) Subscribe(eventName string) Subscription {

	responseChannel := make(chan Subscription, 1)

	b.subscribeChannel <- subscribe{
		eventName: eventName,
		response:  responseChannel,
	}

	return <-responseChannel
}

func (b *MemoryMessageBus) Unsubscribe(subscription Subscription) {
	b.unsubscribeChannel <- subscription
}

func (b *MemoryMessageBus) Close() {
	b.close <- struct{}{}
}

func NewMemoryMessageBus() *MemoryMessageBus {

	eventChannel := make(chan emit, 100)
	subscribeChannel := make(chan subscribe)
	unsubscribeChannel := make(chan Subscription)
	closeChannel := make(chan struct{})

	bus := &MemoryMessageBus{
		eventChannel:       eventChannel,
		subscribeChannel:   subscribeChannel,
		unsubscribeChannel: unsubscribeChannel,
		// sync channel is chosen since we want to wait till the close "signal" is picked up
		close: closeChannel,
	}

	go func() {

		subscriptions := map[string][]chan interface{}{}

		for {

			select {

			case emittedEvent := <-eventChannel:

				// relevant subscriptions for emitted event
				relevantSubscriptions := subscriptions[emittedEvent.eventName]

				// send event to subscriptions
				for _, subscription := range relevantSubscriptions {
					subscription <- emittedEvent.payload
				}

			case subscribe := <-subscribeChannel:
				subscription := make(chan interface{}, 50)
				subscriptions[subscribe.eventName] = append(subscriptions[subscribe.eventName], subscription)
				subscribe.response <- subscription

			case unsubscribe := <-unsubscribeChannel:

				for eventName, localSubscriptions := range subscriptions {

					newSubscriptionSet := []chan interface{}{}

					for _, subscription := range localSubscriptions {

						// in the case the subscription doesn't match the subscription we want to unsubscribe we want to add it to our new
						// subscription set
						if subscription != unsubscribe {
							newSubscriptionSet = append(newSubscriptionSet, subscription)
						}

					}

					// override with new subscription set
					subscriptions[eventName] = newSubscriptionSet

				}
			case <-closeChannel:
				// exit from loop which will close the go routine
				return
			}

		}

	}()

	return bus

}
