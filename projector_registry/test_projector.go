package projector_registry

import "es/event"

type testProjector struct {
	name               string
	interestedInEvents []event.IESEvent
	handleEvent        func(event event.IESEvent) error
}

func (tp *testProjector) Name() string {
	return tp.name
}

func (tp *testProjector) InterestedInEvents() []event.IESEvent {
	return tp.interestedInEvents
}

func (tp *testProjector) Handle(event event.IESEvent) error {
	return tp.handleEvent(event)
}
