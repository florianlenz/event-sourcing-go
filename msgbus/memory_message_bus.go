package msgbus

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
	close              chan chan struct{}
	eventChannel       chan emit
	subscribeChannel   chan subscribe
	unsubscribeChannel chan Subscription
	// DON'T mutate this property outside the go routine
	subscriptions map[string][]chan interface{}
}

func (b *MemoryMessageBus) Emit(eventName string, payload interface{}) {
	b.eventChannel <- emit{
		eventName: eventName,
		payload:   payload,
	}
}

func (b *MemoryMessageBus) Subscribe(eventName string) <-chan interface{} {

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
	closed := make(chan struct{}, 1)
	b.close <- closed
	<-closed
}

func NewMemoryMessageBus() *MemoryMessageBus {

	eventChannel := make(chan emit, 100)
	subscribeChannel := make(chan subscribe)
	unsubscribeChannel := make(chan Subscription)

	// DON'T mutate outside of go routine
	subscriptions := map[string][]chan interface{}{}

	bus := &MemoryMessageBus{
		eventChannel:       eventChannel,
		subscribeChannel:   subscribeChannel,
		unsubscribeChannel: unsubscribeChannel,
		subscriptions:      subscriptions,
	}

	go func() {

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

			}

		}

	}()

	return bus

}
