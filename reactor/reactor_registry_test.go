package reactor

import (
	"github.com/florianlenz/event-sourcing-go/event"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// test reactor with invalid handle method to test certain behaviour
type testReactorWithInvalidHandleMethod struct {
}

func (r *testReactorWithInvalidHandleMethod) Handle(val string) {}

// test reactor with invalid amount of parameters to test the case where we receive an invalid amount of arguments for the handler function
type testReactorWithTooManyParamsInHandleFunction struct {
}

func (r *testReactorWithTooManyParamsInHandleFunction) Handle(firstArg interface{}, secondArg interface{}) {
}

// test event one
type testEventOne struct {
	event.ESEvent
}

// test event two
type testEventTwo struct {
	event.ESEvent
}

// test reactor one
type testReactorOne struct {
	handle func(testEventOne)
}

func (r *testReactorOne) Handle(e testEventOne) {
	r.handle(e)
}

// test reactor two
type testReactorTwo struct {
	handle func(e testEventTwo)
}

func (r *testReactorTwo) Handle(e testEventTwo) {
	r.handle(e)
}

// test reactor three
type testReactorThree struct {
	handle func(e testEventTwo)
}

func (r *testReactorThree) Handle(e testEventTwo) {
	r.handle(e)
}

func TestReactorRegistry(t *testing.T) {

	Convey("reactor registry", t, func() {

		Convey("register", func() {

			Convey("reactor must be a struct", func() {
				rr := NewReactorRegistry()
				reactor := "hi"
				So(rr.Register(&reactor), ShouldBeError, "reactor 'string' must be a struct")
			})

			Convey("reactor must have a Handle method", func() {
				type testReactor struct {
				}
				rr := NewReactorRegistry()
				So(rr.Register(&testReactor{}), ShouldBeError, "reactor 'testReactor' doesn't have a 'Handle' method")
			})

			Convey("reactors Handle method must take exactly one argument", func() {
				rr := NewReactorRegistry()
				So(rr.Register(&testReactorWithTooManyParamsInHandleFunction{}), ShouldBeError, "the handle method of reactor testReactorWithTooManyParamsInHandleFunction must expect exactly one parameter")
			})

			Convey("reactors Handle method must take an implementation of IESEvent  as it's argument", func() {
				rr := NewReactorRegistry()
				So(rr.Register(&testReactorWithInvalidHandleMethod{}), ShouldBeError, "the handle method expects 'string' which is not an IESImplementation")
			})

			Convey("register successfully", func() {
				rr := NewReactorRegistry()
				So(rr.Register(&testReactorOne{}), ShouldBeNil)
			})

			Convey("can't register reactor twice", func() {

				rr := NewReactorRegistry()

				So(rr.Register(&testReactorOne{}), ShouldBeNil)
				So(rr.Register(&testReactorOne{}), ShouldBeError, "reactor 'testReactorOne' has already been registered")
				So(rr.Register(&testReactorOne{}), ShouldBeError, "reactor 'testReactorOne' has already been registered")

			})

		})

		Convey("reactors", func() {

			Convey("fetch reactors", func() {

				rr := NewReactorRegistry()

				// channels where we will receive the signals
				handledReactorOne := false
				handledReactorTwo := false
				handledReactorThree := false

				// register reactors
				So(rr.Register(&testReactorOne{
					handle: func(e testEventOne) {
						handledReactorOne = true
					},
				}), ShouldBeNil)

				So(rr.Register(&testReactorTwo{
					handle: func(e testEventTwo) {
						handledReactorTwo = true
					},
				}), ShouldBeNil)

				So(rr.Register(&testReactorThree{
					handle: func(e testEventTwo) {
						handledReactorThree = true
					},
				}), ShouldBeNil)

				// fetch reactors for test event two
				for _, reactor := range rr.Reactors(testEventOne{}) {
					reactor(testEventOne{})
				}
				So(handledReactorOne, ShouldBeTrue)

				// fetch reactors for event two
				for _, reactor := range rr.Reactors(testEventTwo{}) {
					reactor(testEventTwo{})
				}
				So(handledReactorTwo, ShouldBeTrue)
				So(handledReactorThree, ShouldBeTrue)

			})

			Convey("fetch reactors by pointer event", func() {

				rr := NewReactorRegistry()

				So(rr.Register(&testReactorOne{}), ShouldBeNil)

				reactors := rr.Reactors(&testEventOne{})
				So(reactors, ShouldHaveLength, 1)

			})

		})

	})

}
