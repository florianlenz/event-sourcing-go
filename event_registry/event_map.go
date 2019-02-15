package event_registry

type registeredEvent struct {
	event   event.IEvent
	factory event.FromPayloadFactory
}

type eventMap map[string]event.IEvent

func (r eventMap) AlreadyRegistered(e event.IEvent) bool {
	return false
}
