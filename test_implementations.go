package es

type testLogger struct {
	errorChan chan error
}

func (l *testLogger) Error(error error) {
	l.errorChan <- error
}

type testEvent struct {
	name       string
	payload    Payload
	version    uint8
	occurredAt int64
}

func (e testEvent) Name() string {
	return e.name
}

func (e testEvent) Payload() Payload {
	return e.payload
}

func (e testEvent) Version() uint8 {
	return e.version
}

func (e testEvent) OccurredAt() int64 {
	return e.occurredAt
}

type testProjector struct {
	name               string
	interestedInEvents []IESEvent
	handleEvent        func(event IESEvent) error
}

func (tp *testProjector) Name() string {
	return tp.name
}

func (tp *testProjector) InterestedInEvents() []IESEvent {
	return tp.interestedInEvents
}

func (tp *testProjector) Handle(event IESEvent) error {
	return tp.handleEvent(event)
}
