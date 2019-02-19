package es

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestProjectorRegistry(t *testing.T) {

	Convey("event registry", t, func() {

		Convey("register", func() {

			Convey("should return error on attempt to register different events with the same name", func() {

				registry := New()

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

				registry := New()

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

				registry := New()

				projectors := registry.ProjectorsForEvent(testEvent{
					name: "user.registered",
				})

				So(projectors, ShouldHaveLength, 0)

			})

			Convey("find relevant projectors", func() {

				projectorA := &testProjector{
					name: "projector.a",
					interestedInEvents: []IESEvent{
						testEvent{
							name: "user.registered",
						},
						testEvent{
							name: "user.changed_email",
						},
					},
				}

				projectorB := &testProjector{
					name: "projector.b",
					interestedInEvents: []IESEvent{
						testEvent{
							name: "user.verified_email",
						},
						testEvent{
							name: "user.changed_email",
						},
					},
				}

				projectorC := &testProjector{
					name: "projector.c",
					interestedInEvents: []IESEvent{
						testEvent{
							name: "user.registered",
						},
					},
				}

				registry := New()

				// register projector A
				err := registry.Register(projectorA)
				So(err, ShouldBeNil)

				// register projector B
				err = registry.Register(projectorB)
				So(err, ShouldBeNil)

				// register projector C
				err = registry.Register(projectorC)
				So(err, ShouldBeNil)

				filteredProjectors := registry.ProjectorsForEvent(testEvent{
					name: "user.registered",
				})

				So(filteredProjectors, ShouldContain, projectorA)
				So(filteredProjectors, ShouldContain, projectorC)
				So(filteredProjectors, ShouldNotContain, projectorB)

			})

		})

	})

}
