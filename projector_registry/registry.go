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

		for {

			registeredProjectors := map[string]proj.IProjector{}

			select {
			case addProjector := <-registerProjectorChan:

				// make sure that projector doesn't exist
				_, exists := registeredProjectors[addProjector.projector.Name()]
				if exists {
					addProjector.response <- fmt.Errorf("couldn't add projector since a projector with the same name got already registered")
					continue
				}

				// add projector
				registeredProjectors[addProjector.projector.Name()] = addProjector.projector
				addProjector.response <- nil

			case query := <-queryProjectorsByEvent:

				filteredEvents := []proj.IProjector{}

				for _, projector := range registeredProjectors {
					if projector.InterestedInEvent(query.event) {
						filteredEvents = append(filteredEvents, projector)
					}
				}

				query.responseChan <- filteredEvents

			}

		}

	}()

	return registry

}
