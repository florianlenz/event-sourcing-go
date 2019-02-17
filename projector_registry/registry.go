package projector_registry

import (
	"es/event"
	proj "es/projector"
	"fmt"
)

type addProjector struct {
	projector proj.IProjector
	response  chan error
}

type queryProjectorsByEvent struct {
	event        event.IESEvent
	responseChan chan []proj.IProjector
}

type Registry struct {
	addProjector           chan addProjector
	queryProjectorsByEvent chan queryProjectorsByEvent
}

func (r *Registry) Register(projector proj.IProjector) error {

	respChan := make(chan error, 1)

	r.addProjector <- struct {
		projector proj.IProjector
		response  chan error
	}{projector: projector, response: respChan}

	return <-respChan

}

func (r *Registry) ProjectorsForEvent(event event.IESEvent) []proj.IProjector {

	responseChan := make(chan []proj.IProjector)

	r.queryProjectorsByEvent <- queryProjectorsByEvent{
		event:        event,
		responseChan: responseChan,
	}

	return <-responseChan

}

func New() *Registry {

	registerProjectorChan := make(chan addProjector)

	queryProjectorsByEvent := make(chan queryProjectorsByEvent)

	registry := &Registry{
		addProjector:           registerProjectorChan,
		queryProjectorsByEvent: queryProjectorsByEvent,
	}

	go func() {

		registeredProjectors := map[string]proj.IProjector{}

		for {

			select {
			case addProjector := <-registerProjectorChan:

				responseChannel := addProjector.response
				projectorToAdd := addProjector.projector

				// make sure that projector doesn't exist
				_, exists := registeredProjectors[projectorToAdd.Name()]
				if exists {
					addProjector.response <- fmt.Errorf("projector with name '%s' already registered", projectorToAdd.Name())
					continue
				}

				// add projector
				registeredProjectors[projectorToAdd.Name()] = projectorToAdd
				responseChannel <- nil

			case query := <-queryProjectorsByEvent:

				filteredEvents := []proj.IProjector{}

				for _, projector := range registeredProjectors {
					for _, e := range projector.InterestedInEvents() {
						if e.Name() == query.event.Name() {
							filteredEvents = append(filteredEvents, projector)
						}
					}
				}

				query.responseChan <- filteredEvents

			}

		}

	}()

	return registry

}
