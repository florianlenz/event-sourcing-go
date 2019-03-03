package projector

import (
	"github.com/florianlenz/event-sourcing-go/event"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// test projector
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

// test event
type testEvent struct {
	event.ESEvent
}

func TestProjectorRegistry(t *testing.T) {

	Convey("event registry", t, func() {

		Convey("register", func() {

			Convey("should return error on attempt to register different events with the same name", func() {

				registry := NewProjectorRegistry()

				// register the first time
				err := registry.Register(&testProjector{
					name: "user.projector",
				})
				So(err, ShouldBeNil)

				// register the second time
				err = registry.Register(&testProjector{
					name: "user.projector",
				})
				So(err, ShouldBeError, "projector with id: 'user.projector' has already been registered")

			})

			Convey("register successfully", func() {

				registry := NewProjectorRegistry()

				// register projector
				// @todo it would be nice to do more assertions than just
				err := registry.Register(&testProjector{
					name: "user.projector",
				})
				So(err, ShouldBeNil)

			})

			Convey("shouldn't be able to register event class twice with different name", func() {

			})

		})

		Convey("filter projectors for event", func() {

			Convey("no projectors were registered for event", func() {

				registry := NewProjectorRegistry()

				projectors := registry.ProjectorsForEvent(testEvent{})

				So(projectors, ShouldHaveLength, 0)

			})

			Convey("find relevant projectors", func() {

			})

		})

	})

}
