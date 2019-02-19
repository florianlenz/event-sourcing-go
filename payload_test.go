package es

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPayload(t *testing.T) {

	Convey("Test event payload", t, func() {

		Convey("test get string functionality", func() {

			Convey("should return empty string and error if key is not set", func() {

				payload := Payload{}

				email, err := payload.GetString("e_mail_address")
				So(err, ShouldBeError, "'e_mail_address' does not exist in payload")
				So(email, ShouldEqual, "")

			})

			Convey("should return empty string if value of key is not a valid string", func() {

				payload := Payload{}

				payload["e_mail_address"] = 1

				email, err := payload.GetString("e_mail_address")
				So(err, ShouldBeError, "value of key 'e_mail_address' is not a string")
				So(email, ShouldEqual, "")

			})

			Convey("should return value if value is a string", func() {

				payload := Payload{}
				payload["e_mail_address"] = "florian@test.io"

				email, err := payload.GetString("e_mail_address")
				So(err, ShouldBeNil)
				So(email, ShouldEqual, "florian@test.io")

			})

		})

	})

}
