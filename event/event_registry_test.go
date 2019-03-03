package event

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type testEventPayload struct {
}

type testEvent struct {
	ESEvent
	Payload testEventPayload
}

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
				So(err, ShouldBeError, "user.registered has already been registered")

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

				// try to convert event
				e, err := registry.EventToESEvent(Event{
					Name: "user.created",
				})
				So(err, ShouldBeError, "event 'user.created' hasn't been registered")
				So(e, ShouldBeNil)

			})

			Convey("transform event successfully", func() {

				registry := NewEventRegistry()

				// register event
				err := registry.RegisterEvent("user.created", testEvent{})
				So(err, ShouldBeNil)

				// try to create es event from event
				transformedEvent, err := registry.EventToESEvent(Event{
					Name: "user.created",
				})
				So(err, ShouldBeNil)
				So(transformedEvent, ShouldResemble, testEvent{})

			})

		})

		Convey("get event name", func() {

			Convey("with registered event", func() {

				rr := NewEventRegistry()

				So(rr.RegisterEvent("user.created", testEvent{}), ShouldBeNil)

				eventName, err := rr.GetEventName(&testEvent{})
				So(err, ShouldBeNil)
				So(eventName, ShouldEqual, "user.created")

			})

			Convey("with event that hasn't been registered", func() {

				rr := NewEventRegistry()

				eventName, err := rr.GetEventName(testEvent{})
				So(err, ShouldBeError, "event hasn't been registered")
				So(eventName, ShouldEqual, "")

			})

		})

	})

}
