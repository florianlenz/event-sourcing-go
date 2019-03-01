package reactor

import (
	"github.com/florianlenz/event-sourcing-go/event"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type testReactorWithInvalidHandleMethod struct {
}

func (r *testReactorWithInvalidHandleMethod) Handle(val string) {

}

type testReactorWithTooManyParamsInHandleFunction struct {
}

func (r *testReactorWithTooManyParamsInHandleFunction) Handle(firstArg interface{}, secondArg interface{}) {

}

type testEvent struct {
	event.ESEvent
}

type testReactor struct {
}

func (r *testReactor) Handle(e testEvent) {

}

type anotherTestReactor struct {
}

func (r *anotherTestReactor) Handle(e testEvent) {

}

func TestReactorRegistry(t *testing.T) {

	Convey("reactor registry", t, func() {

		Convey("register", func() {

			Convey("reactor must be a struct", func() {
				rr := NewReactorRegistry()
				So(rr.Register("hi"), ShouldBeError, "reactor 'string' must be a struct")
			})

			Convey("reactor must have a Handle method", func() {
				type testReactor struct {
				}
				rr := NewReactorRegistry()
				So(rr.Register(testReactor{}), ShouldBeError, "reactor 'testReactor' doesn't have a 'Handle' method")
			})

			Convey("reactors Handle method must take exactly one argument", func() {
				rr := NewReactorRegistry()
				So(rr.Register(testReactorWithTooManyParamsInHandleFunction{}), ShouldBeError, "the handle method of reactor testReactorWithTooManyParamsInHandleFunction must expect exactly one parameter")
			})

			Convey("reactors Handle method must take an implementation of IESEvent  as it's argument", func() {
				rr := NewReactorRegistry()
				So(rr.Register(testReactorWithInvalidHandleMethod{}), ShouldBeError, "the handle method expects 'string' which is not an IESImplementation")
			})

			Convey("register successfully", func() {
				rr := NewReactorRegistry()
				So(rr.Register(testReactor{}), ShouldBeNil)
			})

			Convey("can't register reactor twice", func() {

				rr := NewReactorRegistry()

				So(rr.Register(testReactor{}), ShouldBeNil)
				So(rr.Register(testReactor{}), ShouldBeError, "reactor 'testReactor' has already been registered")
				So(rr.Register(&testReactor{}), ShouldBeNil, "reactor 'testReactor' has already been registered")

			})

		})

	})

}
