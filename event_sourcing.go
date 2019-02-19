package es

import (
	"errors"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"time"
)

type addProcessedListener struct {
	listener chan struct{}
	eventID  primitive.ObjectID
}

type EventSourcing struct {
	eventRepository          IEventRepository
	messageBus               IMessageBus
	addProcessedListenerChan chan addProcessedListener
	close                    chan struct{}
}

func (es *EventSourcing) Commit(e IESEvent, onProcessed chan struct{}) error {

	// new event
	eventToPersist := &event{
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
			eventID:  *eventToPersist.ID,
		}
	}

	// send event to message bus
	es.messageBus.Emit("event:occurred", eventToPersist.ID)

	return nil

}

func NewEventSourcing(eventRepository IEventRepository, mb IMessageBus, logger ILogger) *EventSourcing {

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
		processedEventListeners := map[string]chan struct{}{}

		for {

			select {

			// handle processed events
			case value := <-subscriber:

				// cast to event
				eventID, k := value.(primitive.ObjectID)
				if !k {
					logger.Error(errors.New("didn't receive an event id"))
					continue
				}

				// event
				event, err := eventRepository.FetchByID(eventID)
				if err != nil {
					logger.Error(err)
					continue
				}

				// fetch listener
				listener, exists := processedEventListeners[eventID.Hex()]
				if exists {
					delete(processedEventListeners, event.ID.Hex())
					// notify listener that the event got processed
					listener <- struct{}{}
				}

			// add new processed event listener
			case eventProcessedListener := <-addProcessedListenerChan:

				// make sure listener was not added before
				_, exists := processedEventListeners[eventProcessedListener.eventID.Hex()]
				if exists {
					logger.Error(fmt.Errorf("listener for event id '%s' already added", eventProcessedListener.eventID))
					continue
				}

				// add listener
				processedEventListeners[eventProcessedListener.eventID.Hex()] = eventProcessedListener.listener

			// break the loop and kill to go routine
			case <-closeChan:
				break
			}

		}

	}()

	return es

}
