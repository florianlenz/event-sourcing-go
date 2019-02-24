package es

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestReactorRegistry(t *testing.T) {

	Convey("reactor registry", t, func() {

		rr := NewReactorRegistry()

		reactor := &testReactor{
			onEvent: "user.created",
		}

		// no reactors were registered - should be empty
		So(rr.ForEvent(testEvent{name: "user.created"}), ShouldBeEmpty)

		// register reactor
		rr.Register(reactor)

		// Should contain registered reactor
		So(rr.ForEvent(testEvent{name: "user.created"}), ShouldContain, reactor)

	})

}
