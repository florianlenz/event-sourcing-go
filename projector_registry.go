package es

import "fmt"

type addProjector struct {
	projector IProjector
	response  chan error
}

type queryProjectorsByEvent struct {
	event        IESEvent
	responseChan chan []IProjector
}

type projectorRegistry struct {
	addProjector           chan addProjector
	queryProjectorsByEvent chan queryProjectorsByEvent
}

func (r *projectorRegistry) Register(projector IProjector) error {

	respChan := make(chan error, 1)

	r.addProjector <- struct {
		projector IProjector
		response  chan error
	}{projector: projector, response: respChan}

	return <-respChan

}

func (r *projectorRegistry) ProjectorsForEvent(event IESEvent) []IProjector {

	responseChan := make(chan []IProjector)

	r.queryProjectorsByEvent <- queryProjectorsByEvent{
		event:        event,
		responseChan: responseChan,
	}

	return <-responseChan

}

func newProjectorRegistry() *projectorRegistry {

	registerProjectorChan := make(chan addProjector)

	queryProjectorsByEvent := make(chan queryProjectorsByEvent)

	registry := &projectorRegistry{
		addProjector:           registerProjectorChan,
		queryProjectorsByEvent: queryProjectorsByEvent,
	}

	go func() {

		registeredProjectors := map[string]IProjector{}

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

				filteredEvents := []IProjector{}

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
