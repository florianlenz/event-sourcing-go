package es

type IReactor interface {
	// react on event
	Handle(event IESEvent)
	OnEvent() string
}

type ReactorRegistry struct {
	registerReactorChannel       chan IReactor
	fetchReactorsForEventChannel chan fetchReactorsForEvent
}

type registerReactor struct {
	eventName string
	reactor   IReactor
}

type fetchReactorsForEvent struct {
	eventName string
	response  chan []IReactor
}

// Register a new reactor
func (r *ReactorRegistry) Register(reactor IReactor) {
	r.registerReactorChannel <- reactor
}

// Fetch reactors for event
func (r *ReactorRegistry) ForEvent(event IESEvent) []IReactor {

	response := make(chan []IReactor, 1)

	r.fetchReactorsForEventChannel <- fetchReactorsForEvent{
		eventName: event.Name(),
		response:  response,
	}

	return <-response

}

func NewReactorRegistry() *ReactorRegistry {

	registerReactorChannel := make(chan IReactor)
	fetchReactorsForEventChannel := make(chan fetchReactorsForEvent)

	r := &ReactorRegistry{
		registerReactorChannel:       registerReactorChannel,
		fetchReactorsForEventChannel: fetchReactorsForEventChannel,
	}

	go func() {

		registeredReactors := map[string][]IReactor{}

		for {

			select {
			case reactor := <-registerReactorChannel:
				// register reactor
				registeredReactors[reactor.OnEvent()] = append(registeredReactors[reactor.OnEvent()], reactor)
			case fetch := <-fetchReactorsForEventChannel:
				// fetch reactors for event
				fetch.response <- registeredReactors[fetch.eventName]
			}

		}

	}()

	return r

}
