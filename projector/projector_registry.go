package projector

import (
	"fmt"
	"github.com/florianlenz/event-sourcing-go/event"
	"reflect"
	"sync"
)

type Registry struct {
	lock       *sync.Mutex
	projectors map[string]IProjector
}

// register an projector
func (r *Registry) Register(projector IProjector) error {

	// lock
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
	}()

	// ensure that projector hasn't been added
	_, exists := r.projectors[projector.Name()]
	if exists {
		return fmt.Errorf("projector with id: '%s' has already been registered", projector.Name())
	}

	r.projectors[projector.Name()] = projector

	return nil

}

func (r *Registry) ProjectorsForEvent(event event.IESEvent) []IProjector {

	// lock
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
	}()

	// event type
	eventType := reflect.TypeOf(event)
	if eventType.Kind() == reflect.Ptr {
		eventType = eventType.Elem()
	}

	projectors := []IProjector{}

	for _, proj := range r.projectors {

		for _, e := range proj.InterestedInEvents() {

			// event type
			eType := reflect.TypeOf(e)
			if eType.Kind() == reflect.Ptr {
				eType = eType.Elem()
			}

			if eType == eventType {
				projectors = append(projectors, proj)
			}

		}

	}

	return projectors

}

func NewProjectorRegistry() *Registry {

	return &Registry{
		lock:       &sync.Mutex{},
		projectors: map[string]IProjector{},
	}

}
