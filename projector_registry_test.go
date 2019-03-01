package es

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

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
				So(err, ShouldBeError, "projector with name 'user.projector' already registered")

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
