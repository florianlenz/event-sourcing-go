package consistent

import (
	"es"
	"es/projector"
	"github.com/jinzhu/gorm"
)

type ProjectorRepository struct {
	db *gorm.DB
}

func (r *ProjectorRepository) outOfSyncQuery(projector Projector) *gorm.DB {
	return r.db.Where("id > ? AND name IN ?", projector.LastProcessedEvent, projector.ProjectorID)
}

func (r *ProjectorRepository) OutOfSyncBy(projector Projector) (uint, error) {
	var count uint
	err := r.outOfSyncQuery(projector).Count(&count).Error
	return count, err
}

func (r *ProjectorRepository) GetOrCreateProjector(projector projector.IProjector) (*Projector, error) {
	fetchedProjector := &Projector{}
	err := r.db.Where("projector_id = ?", fetchedProjector).First(fetchedProjector).Error
	return fetchedProjector, err
}

func (r *ProjectorRepository) Save(projector *Projector) error {
	return r.db.Save(projector).Error
}

func (r *ProjectorRepository) UpdateLastHandledEvent(projector *Projector, event es.Event) error {
	projector.LastProcessedEvent = event.ID
	return r.db.Save(projector).Error
}
