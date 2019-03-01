package es

import (
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func TestUtils(t *testing.T) {

	Convey("utils", t, func() {

		Convey("event payload type", func() {

			Convey("get correct event payload type", func() {

				type payload struct {
				}

				type testEvent struct {
					ESEvent
					payload payload
				}

				payloadType, err := eventPayloadType(testEvent{})
				So(err, ShouldBeNil)
				So(payloadType, ShouldEqual, reflect.TypeOf(payload{}))

			})

			Convey("pointer event is accepted", func() {

				type payload struct {
				}

				type testEvent struct {
					ESEvent
					payload payload
				}

				// test with pointer
				payloadType, err := eventPayloadType(&testEvent{})
				So(err, ShouldBeNil)
				So(payloadType, ShouldEqual, reflect.TypeOf(payload{}))

			})

			Convey("payload does not exists", func() {

				type testEvent struct {
					ESEvent
				}

				_, err := eventPayloadType(&testEvent{})
				So(err, ShouldBeError, "event: 'testEvent' doesn't have a payload property - keep in mind that exported payload properties are not accepted")

			})

			Convey("payload must be a struct", func() {

				type testEvent struct {
					ESEvent
					payload string
				}

				_, err := eventPayloadType(&testEvent{})
				So(err, ShouldBeError, "the payload of event 'testEvent' must be a struct - got: 'string'")

			})

		})

		Convey("does event has valid factory method", func() {

			Convey("has valid factory method (with pointer passed as event reference)", func() {

			})

			Convey("has valid factory method", func() {

			})

			Convey("exit if factory method doesn't exist", func() {

			})

			Convey("factory method must expect two parameters", func() {

			})

			// @todo is argument correct for the returned value?
			Convey("factory method must return exactly one argument", func() {

			})

			Convey("the first expected parameter of the factory method must be of type ESEvent", func() {

			})

			Convey("the second expected parameter of the factory method must be of the payload type", func() {

			})

			Convey("the returned argument must be the same type as the event", func() {

			})

		})

		Convey("reactor factory", func() {

			Convey("reactor must be a struct", func() {

			})

			Convey("reactor must have Handle method", func() {

			})

			Convey("reactors handle method must expect exactly one argument", func() {

			})

			Convey("reactors handle method must except a struct that implements the IESEvent interface", func() {

			})

		})

	})

}
