package es

import (
	e "es/event"
	mb "es/msgbus"
	"time"
)

type addProcessedListener struct {
	listener chan struct{}
	eventID  uint64
}

type EventSourcing struct {
	eventRepository          IEventRepository
	messageBus               mb.IMessageBus
	addProcessedListenerChan chan addProcessedListener
	close                    chan struct{}
}

func (es *EventSourcing) Commit(e e.IESEvent, onProcessed chan struct{}) error {

	// new event
	eventToPersist := &Event{
		Name:       e.Name(),
		Payload:    e.Payload(),
		Version:    e.Version(),
		OccurredAt: time.Now().Unix(),
	}

	// persist event
	if err := es.eventRepository.Save(eventToPersist); err != nil {
		return err
	}

	// add processed listener
	if onProcessed != nil {
		es.addProcessedListenerChan <- addProcessedListener{
			listener: onProcessed,
			eventID:  eventToPersist.ID,
		}
	}

	// send event to message bus
	es.messageBus.Emit("event:occurred", eventToPersist.ID)

	return nil

}

func New(eventRepository IEventRepository, mb mb.IMessageBus) *EventSourcing {

	addProcessedListenerChan := make(chan addProcessedListener)
	closeChan := make(chan struct{})

	es := &EventSourcing{
		eventRepository:          eventRepository,
		messageBus:               mb,
		addProcessedListenerChan: addProcessedListenerChan,
		close:                    closeChan,
	}

	// event store state machine
	go func() {

		// subscribe to processed events
		subscriber := mb.Subscribe("event:processed")

		// processed event listeners
		processedEventListeners := map[uint64]chan struct{}{}

		for {

			select {

			// handle processed events
			case value := <-subscriber:

				// cast to event
				event, k := value.(Event)
				if !k {
					// @todo log
					panic("event is not an persisted event")
				}

				// fetch listener
				listener, exists := processedEventListeners[event.ID]
				if exists {
					processedEventListeners[event.ID] = nil
					// notify listener that the event got processed
					listener <- struct{}{}
				}

			// add new processed event listener
			case eventProcessedListener := <-addProcessedListenerChan:

				// make sure listener was not added before
				_, exists := processedEventListeners[eventProcessedListener.eventID]
				if exists {
					panic("event listener already added for event id - this shouldn't happen")
				}

				// add listener
				processedEventListeners[eventProcessedListener.eventID] = eventProcessedListener.listener

			// break the loop and kill to go routine
			case <-closeChan:
				break
			}

		}

	}()

	return es

}
