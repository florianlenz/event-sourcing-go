package es

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestMemeoryMessageBus(t *testing.T) {

	Convey("memory message bus", t, func() {

		Convey("subscribe", func() {

			// message bus
			bus := NewMemoryMessageBus()

			// make sure that there are no subscriptions
			So(bus.subscriptions["my:event"], ShouldHaveLength, 0)

			// subscribe
			bus.Subscribe("my:event")

			// there should be one subscription for my:event
			So(bus.subscriptions["my:event"], ShouldHaveLength, 1)

		})

		Convey("unsubscribe", func() {

			// message bus
			bus := NewMemoryMessageBus()

			// subscribe
			subscription := bus.Subscribe("my:event")

			// ensure the subscription got added
			So(bus.subscriptions["my:event"], ShouldHaveLength, 1)

			// unsubscribe
			bus.Unsubscribe(subscription)

			// wait a bit till the go routine is synchronized
			time.Sleep(time.Second)

			// ensure the subscription got removed
			So(bus.subscriptions["my:event"], ShouldHaveLength, 0)

		})

		Convey("emit event", func() {

			// message bus
			bus := NewMemoryMessageBus()

			// subscriptions
			firstSubscription := bus.Subscribe("my:event")
			secondSubscription := bus.Subscribe("my:event")
			thirdSubscription := bus.Subscribe("another:event")

			// send event to bus
			bus.Emit("my:event", "payload")

			// first and second subscriptions must receive payload
			So(<-firstSubscription, ShouldEqual, "payload")
			So(<-secondSubscription, ShouldEqual, "payload")

			// make sure that the third subscription (subscribed for other event) don't receive a payload
			receivedOnThirdSubscription := false
			select {
			case <-thirdSubscription:
				receivedOnThirdSubscription = true
			case <-time.After(time.Second):
			}
			So(receivedOnThirdSubscription, ShouldBeFalse)

		})

	})

}
