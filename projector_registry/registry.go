package projector_registry

import (
	"es/event"
	"es/projector"
	"fmt"
)

type addProjector struct {
	projector projector.IProjector
	response  chan error
}

type queryProjectorsByEvent struct {
	event        event.IESEvent
	responseChan chan []projector.IProjector
}

type Registry struct {
	addProjector           chan addProjector
	queryProjectorsByEvent chan queryProjectorsByEvent
}

func (r *Registry) Register(projector projector.IProjector) {

}

func (r *Registry) ProjectorsForEvent(eventName string) []projector.IProjector {

}

func New() *Registry {

	registerProjectorChan := make(chan addProjector)

	queryProjectorsByEvent := make(chan queryProjectorsByEvent)

	registry := &Registry{
		addProjector:           registerProjectorChan,
		queryProjectorsByEvent: queryProjectorsByEvent,
	}

	go func() {

		for {

			registeredProjectors := map[string]projector.IProjector{}

			select {
			case addProjector := <-registerProjectorChan:

				// make sure that projector doesn't exist
				_, exists := registeredProjectors[addProjector.projector.Name()]
				if exists {
					addProjector.response <- fmt.Errorf("couldn't add projector since a projector with the same name got already registered")
				}

				// add projector
				registeredProjectors[addProjector.projector.Name()] = addProjector.projector
				addProjector.response <- nil

			case query := <-queryProjectorsByEvent:

				filteredEvents := []projector.IProjector{}

				for _, proje := range registeredProjectors {
					if proje.InterestedInEvent(query.event) {
						filteredEvents = append(filteredEvents, proje)
					}
				}

				query.responseChan <- filteredEvents

			}

		}

	}()

	return registry

}
