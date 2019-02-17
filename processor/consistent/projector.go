package consistent

import (
	"es"
	"es/projector"
	"github.com/jinzhu/gorm"
)

type IProjectorRepository interface {
	// check if projector is out of sync
	OutOfSyncBy(projector Projector) (uint, error)
	// update the last handled event on the projector
	UpdateLastHandledEvent(projector *Projector, event es.Event) error
	// fetch projector
	GetOrCreateProjector(projector projector.IProjector) (*Projector, error)
	// persist projector
	Save(projector *Projector) error
}

// projector entry in the database
type Projector struct {
	gorm.Model
	ProjectorID        string `gorm:"unique"`
	LastProcessedEvent uint64
}
