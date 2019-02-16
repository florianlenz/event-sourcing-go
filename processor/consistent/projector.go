package consistent

import (
	"es/projector"
	"es/store"
	"github.com/jinzhu/gorm"
)

type IProjectorRepository interface {
	// check if projector is out of sync
	OutOfSync(projector projector.IProjector) bool
	// update the last handled event on the projector
	UpdateLastHandledEvent(projector projector.IProjector, event store.Event)
}

// projector entry in the database
type Projector struct {
	gorm.Model
	ProjectorID        string `gorm:"unique"`
	LastProcessedEvent uint64
}
