package event

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSpec(t *testing.T) {

	Convey("event registry", t, func() {

		Convey("register event", func() {

			Convey("should return error on attempt to register an event with the same name twice", func() {

				registry := NewEventRegistry()

				// first attempt
				err := registry.RegisterEvent("user.registered", testEvent{})
				So(err, ShouldBeNil)

				// second attempt
				err = registry.RegisterEvent("user.registered", testEvent{})
				So(err, ShouldBeError, "event with name 'user.registered' got already registered")

			})

			Convey("register event successfully", func() {

				registry := NewEventRegistry()

				// register event
				err := registry.RegisterEvent("user.registered", testEvent{})
				So(err, ShouldBeNil)

			})

		})

		Convey("test event to es event transformation", func() {

			Convey("should return an error if the event hasn't been registered", func() {

				registry := NewEventRegistry()

				// register event
				e, err := registry.EventToESEvent(event{
					Name: "user.created",
				})
				So(err, ShouldBeError, "event with name 'user.created' hasn't been registered")
				So(e, ShouldBeNil)

			})

			Convey("should return error if the transformed event is has a different name than the persisted event", func() {

				registry := NewEventRegistry()

				// register event
				err := registry.RegisterEvent("user.created", testEvent{})
				So(err, ShouldBeNil)

				// try to create es event from event
				esEvent, err := registry.EventToESEvent(event{
					Name: "user.created",
				})
				So(err, ShouldBeError, "attention! the creation of an event with name 'user.created' resulted in the creation of an event with name: 'wrong.event_name'")
				So(esEvent, ShouldBeNil)

			})

			Convey("transform event successfully", func() {

				registry := NewEventRegistry()

				// test event
				esEvent := &testEvent{}

				// register event
				err := registry.RegisterEvent("user.created", testEvent{})
				So(err, ShouldBeNil)

				// try to create es event from event
				transformedEvent, err := registry.EventToESEvent(event{
					Name: "user.created",
				})
				So(err, ShouldBeNil)
				So(transformedEvent, ShouldEqual, esEvent)

			})

		})

	})

}
