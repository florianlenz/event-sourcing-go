package event

import (
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

type testUsername struct {
	username string
}

func (u *testUsername) Marshal() (interface{}, error) {
	return u.username, nil
}

func (u *testUsername) Unmarshal(param interface{}) error {

	un, k := param.(string)
	if !k {
		return errors.New("expected string")
	}

	u.username = un

	return nil
}

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

		Convey("event payload to map", func() {

			Convey("types", func() {

				Convey("String", func() {

					type Payload struct {
						Name string `es:"name"`
					}

					type event struct {
						ESEvent
						Payload Payload
					}

					marshaledContent, err := PayloadToMap(event{
						Payload: Payload{
							Name: "Jens",
						},
					})
					So(err, ShouldBeNil)
					So(marshaledContent, ShouldResemble, map[string]interface{}{
						"name": "Jens",
					})

				})

				Convey("Bool", func() {

					type Payload struct {
						Created bool `es:"created"`
					}

					type event struct {
						ESEvent
						Payload Payload
					}

					marshaledContent, err := PayloadToMap(event{
						Payload: Payload{
							Created: false,
						},
					})
					So(err, ShouldBeNil)
					So(marshaledContent, ShouldResemble, map[string]interface{}{
						"created": false,
					})

				})

				Convey("Int", func() {

					type Payloadd struct {
						Age int `es:"age"`
					}

					type event struct {
						ESEvent
						Payload Payloadd
					}

					marshaledContent, err := PayloadToMap(event{
						Payload: Payloadd{
							Age: 5,
						},
					})
					So(err, ShouldBeNil)
					So(marshaledContent, ShouldResemble, map[string]interface{}{
						"age": 5,
					})

				})

				Convey("Uint", func() {

					type Payload struct {
						Age uint `es:"age"`
					}

					type event struct {
						ESEvent
						Payload Payload
					}

					marshaledContent, err := PayloadToMap(event{
						Payload: Payload{
							Age: 39,
						},
					})
					So(err, ShouldBeNil)
					So(marshaledContent["age"], ShouldEqual, 39)

				})

				Convey("Float32", func() {

					type Payload struct {
						Weight float32 `es:"weight"`
					}

					type event struct {
						ESEvent
						Payload Payload
					}

					marshaledContent, err := PayloadToMap(event{
						Payload: Payload{
							Weight: 3.2,
						},
					})
					So(err, ShouldBeNil)
					So(marshaledContent["weight"], ShouldAlmostEqual, float64(3.2), .0000001)

				})

				Convey("Float64", func() {

					type Payload struct {
						Weight float64 `es:"weight"`
					}

					type event struct {
						ESEvent
						Payload Payload
					}

					marshaledContent, err := PayloadToMap(event{
						Payload: Payload{
							Weight: 39399939.333,
						},
					})
					So(err, ShouldBeNil)
					So(marshaledContent, ShouldResemble, map[string]interface{}{
						"weight": 39399939.333,
					})

				})

				Convey("Struct", func() {

					Convey("struct that doesn't support the Marshal interface", func() {

					})

					Convey("marshal successful", func() {

						type Payload struct {
							Username testUsername `es:"username"`
						}

						type event struct {
							ESEvent
							Payload Payload
						}

						marshaledContent, err := PayloadToMap(event{
							Payload: Payload{
								Username: testUsername{
									username: "hans_peter",
								},
							},
						})
						So(err, ShouldBeNil)
						So(marshaledContent, ShouldResemble, map[string]interface{}{
							"username": "hans_peter",
						})

					})

				})

			})

			Convey("pointer to event", func() {

				type Payload struct {
					Name string `es:"name"`
				}

				type event struct {
					ESEvent
					Payload Payload
				}

				marshaledContent, err := PayloadToMap(&event{
					Payload: Payload{
						Name: "Jens",
					},
				})
				So(err, ShouldBeNil)
				So(marshaledContent, ShouldResemble, map[string]interface{}{
					"name": "Jens",
				})

			})

			Convey("missing 'es' tag in payload struct", func() {

				type Payload struct {
					Name string
				}

				type event struct {
					ESEvent
					Payload Payload
				}

				marshaledContent, err := PayloadToMap(event{
					Payload: Payload{
						Name: "Jens",
					},
				})
				So(err, ShouldBeError, "missing 'es' tag in events payload field (event: 'event', payload field: 'Name')")
				So(marshaledContent, ShouldBeNil)

			})

			Convey("unexported payload field", func() {

			})

		})

		Convey("payload map to payload", func() {

			Convey("types", func() {

				Convey("String", func() {

					type Payload struct {
						Name string `es:"name"`
					}

					type event struct {
						ESEvent
						Payload Payload
					}

					e, err := createIESEvent(event{}, Event{
						Payload: map[string]interface{}{
							"name": "Hans",
						},
					})
					So(err, ShouldBeNil)
					esEvent := e.(event)

					So(esEvent.Payload.Name, ShouldEqual, "Hans")

				})

				Convey("Bool", func() {

					type Payload struct {
						Created bool `es:"created"`
					}

					type event struct {
						ESEvent
						Payload Payload
					}

					e, err := createIESEvent(event{}, Event{
						Payload: map[string]interface{}{
							"created": false,
						},
					})
					So(err, ShouldBeNil)
					esEvent := e.(event)

					So(esEvent.Payload.Created, ShouldBeFalse)

				})

				Convey("Int", func() {

					type Payload struct {
						Age int `es:"age"`
					}

					type event struct {
						ESEvent
						Payload Payload
					}

					// create event from payload
					e, err := createIESEvent(event{}, Event{
						Payload: map[string]interface{}{
							"age": 55,
						},
					})
					So(err, ShouldBeNil)
					esEvent := e.(event)

					So(esEvent.Payload.Age, ShouldEqual, 55)

				})

				Convey("Uint", func() {

					type Payload struct {
						Age uint `es:"age"`
					}

					type event struct {
						ESEvent
						Payload Payload
					}

					// create event from payload
					e, err := createIESEvent(event{}, Event{
						Payload: map[string]interface{}{
							"age": uint(22),
						},
					})
					So(err, ShouldBeNil)
					esEvent := e.(event)

					So(esEvent.Payload.Age, ShouldEqual, 22)

				})

				Convey("Float32", func() {

					type Payload struct {
						Price float32 `es:"price"`
					}

					type event struct {
						ESEvent
						Payload Payload
					}

					// create event from payload
					e, err := createIESEvent(event{}, Event{
						Payload: map[string]interface{}{
							"price": 10.30,
						},
					})
					So(err, ShouldBeNil)
					esEvent := e.(event)

					So(esEvent.Payload.Price, ShouldAlmostEqual, 10.30, .000001)

				})

				Convey("Float64", func() {

					type Payload struct {
						Price float64 `es:"price"`
					}

					type event struct {
						ESEvent
						Payload Payload
					}

					// create event from payload
					e, err := createIESEvent(event{}, Event{
						Payload: map[string]interface{}{
							"price": 11.30,
						},
					})
					So(err, ShouldBeNil)
					esEvent := e.(event)

					So(esEvent.Payload.Price, ShouldAlmostEqual, 11.30, .00000001)

				})

				Convey("Struct", func() {

					Convey("unmarshal non pointer successfully", func() {

						type Payload struct {
							Username testUsername `es:"username"`
						}

						type event struct {
							ESEvent
							Payload Payload
						}

						e, err := createIESEvent(event{}, Event{
							Payload: map[string]interface{}{
								"username": "hans_peter",
							},
						})
						So(err, ShouldBeNil)

						esEvent := e.(event)
						So(esEvent.Payload.Username.username, ShouldEqual, "hans_peter")

					})

				})

			})

		})

		Convey("create IESEvent", func() {

			Convey("recover and attach ESEvent to IESEvent instance", func() {

				type testEvent struct {
					ESEvent
					Payload struct {
					}
				}

				e, err := createIESEvent(testEvent{}, Event{
					OccurredAt: 333,
					Version:    1,
				})
				So(err, ShouldBeNil)

				So(e.Version(), ShouldEqual, 1)
				So(e.OccurredAt(), ShouldEqual, 333)

			})

		})

	})

}
