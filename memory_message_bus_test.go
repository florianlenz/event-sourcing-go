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
			subscription := bus.Subscribe("my:event")

			// emit event
			bus.Emit("my:event", "payload")

			// should receive payload if subscription is setup properly
			So(<-subscription, ShouldEqual, "payload")

		})

		Convey("unsubscribe", func() {

			// message bus
			bus := NewMemoryMessageBus()

			// subscribe
			firstSubscription := bus.Subscribe("my:event")
			secondSubscription := bus.Subscribe("my:event")

			// emit test event
			bus.Emit("my:event", "payload")

			// ensure the subscriptions got added by sending events to them
			So(<-firstSubscription, ShouldEqual, "payload")
			So(<-secondSubscription, ShouldEqual, "payload")

			// unsubscribe first subscription
			bus.Unsubscribe(firstSubscription)

			bus.Emit("my:event", "another-payload")

			So(<-secondSubscription, ShouldEqual, "another-payload")

			receivedFromFirstSubscription := false
			select {
			case <-firstSubscription:
				receivedFromFirstSubscription = true
			case <-time.After(time.Second):
			}

			// we should not receive something from the first subscription since unsubscribed from it
			So(receivedFromFirstSubscription, ShouldBeFalse)

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

		Convey("close", func() {

			bus := NewMemoryMessageBus()

			// subscribe
			subscription := bus.Subscribe("my:event")

			// emit event
			bus.Emit("my:event", "payload")

			// make sure that subscription is notified
			So(<-subscription, ShouldEqual, "payload")

			// close the message bus
			bus.Close()

			// send another event
			bus.Emit("my:event", "another:payload")

			receivedFromSubscription := false

			select {
			case <-subscription:
				receivedFromSubscription = true
			case <-time.After(time.Second):
			}

			// should have false since we closed to go routine, which means that no events are passed down anymore to the subscriptions
			So(receivedFromSubscription, ShouldBeFalse)

		})

	})

}
