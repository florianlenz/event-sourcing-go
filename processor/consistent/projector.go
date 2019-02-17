package consistent

import (
	"es"
	"es/projector"
	"github.com/jinzhu/gorm"
)

type IProjectorRepository interface {
	// check if projector is out of sync
	OutOfSync(projector projector.IProjector) bool
	// update the last handled event on the projector
	UpdateLastHandledEvent(projector projector.IProjector, event es.Event)
}

// projector entry in the database
type Projector struct {
	gorm.Model
	ProjectorID        string `gorm:"unique"`
	LastProcessedEvent uint64
}
