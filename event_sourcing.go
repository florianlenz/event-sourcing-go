package es

import (
	"github.com/mongodb/mongo-go-driver/mongo"
	"time"
)

type EventSourcing struct {
	eventRepository iEventRepository
	close           chan struct{}
	processor       *Processor
	eventRegistry   *EventRegistry
}

func (es *EventSourcing) Commit(e IESEvent, onProcessed chan struct{}) error {

	// @todo fetch event name based on type
	eventName, err := es.eventRegistry.GetEventName(e)
	if err != nil {
		return err
	}

	// @todo marshal event payload
	eventPayload, err := marshalEventPayload()
	if err != nil {
		return err
	}

	// new event
	eventToPersist := &event{
		Name:       eventName,
		Payload:    eventPayload,
		Version:    e.Version(),
		OccurredAt: time.Now().Unix(),
	}

	// persist event
	if err := es.eventRepository.Save(eventToPersist); err != nil {
		return err
	}

	// processed
	processedChan := es.processor.Process(*eventToPersist.ID)

	if onProcessed != nil {
		go func() {
			// wait till event got processed
			<-processedChan
			// send processed signal to the passed onProcessed channel
			onProcessed <- struct{}{}
		}()
	}

	return nil

}

func (es *EventSourcing) Start() {
	es.processor.Start()
}

// create a new event sourcing instance. Don't forget to start it. The processor won't process till you told him to do so.
func NewEventSourcing(logger ILogger, db *mongo.Database, projectorRegistry *ProjectorRegistry, eventRegistry *EventRegistry, reactorRegistry *ReactorRegistry) *EventSourcing {

	closeChan := make(chan struct{})

	// collections
	eventCollection := db.Collection("events")
	projectorCollection := db.Collection("projectors")

	// repos
	eventRepository := newEventRepository(eventCollection)
	projectorRepository := newProjectorRepository(eventCollection, projectorCollection, eventRegistry)

	// processor
	processor := newProcessor(projectorRegistry, eventRegistry, reactorRegistry, projectorRepository, eventRepository, logger, false)

	es := &EventSourcing{
		eventRepository: eventRepository,
		close:           closeChan,
		processor:       processor,
		eventRegistry:   eventRegistry,
	}

	return es

	/**
	This was ment to be used in case that there is a message bus that is communicating between the processor and the emitted events.
	This is right now not the case so it's commented. When we bring back the message bus we can use this again

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
				return
			}

		}

	}()

	*/

}
