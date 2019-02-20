package es

import "github.com/mongodb/mongo-go-driver/bson/primitive"

// test logger
type testLogger struct {
	errorChan chan error
}

func (l *testLogger) Error(error error) {
	l.errorChan <- error
}

// test event
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

// test projector
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

// test event repository
type testEventRepository struct {
	save      func(event *event) error
	fetchByID func(id primitive.ObjectID) (event, error)
}

func (r *testEventRepository) Save(event *event) error {
	return r.save(event)
}

func (r *testEventRepository) FetchByID(id primitive.ObjectID) (event, error) {
	return r.fetchByID(id)
}

// test projector repository
type testProjectorRepository struct {
	outOfSyncBy            func(projector IProjector) (int64, error)
	updateLastHandledEvent func(projector IProjector, event event) error
}

func (r *testProjectorRepository) OutOfSyncBy(projector IProjector) (int64, error) {
	return r.outOfSyncBy(projector)
}

func (r *testProjectorRepository) UpdateLastHandledEvent(projector IProjector, event event) error {
	return r.updateLastHandledEvent(projector, event)
}
