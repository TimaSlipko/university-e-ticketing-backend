// internal/repositories/event_repository.go
package repositories

import (
	"eticketing/internal/models"
	"gorm.io/gorm"
)

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) GetByID(id uint) (*models.Event, error) {
	var event models.Event
	err := r.db.Preload("Seller").Preload("Tickets").First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) Update(event *models.Event) error {
	return r.db.Save(event).Error
}

func (r *eventRepository) Delete(id uint) error {
	return r.db.Delete(&models.Event{}, id).Error
}

func (r *eventRepository) ListByStatus(status models.EventStatus, limit, offset int) ([]models.Event, error) {
	var events []models.Event
	err := r.db.Preload("Seller").Where("status = ?", status).Order("date").Limit(limit).Offset(offset).Find(&events).Error
	return events, err
}

func (r *eventRepository) ListByStatusReverse(status models.EventStatus, limit, offset int) ([]models.Event, error) {
	var events []models.Event
	err := r.db.Preload("Seller").Where("status = ?", status).Order("id DESC").Limit(limit).Offset(offset).Find(&events).Error
	return events, err
}

func (r *eventRepository) ListBySeller(sellerID uint, limit, offset int) ([]models.Event, error) {
	var events []models.Event
	err := r.db.Preload("Seller").Where("seller_id = ?", sellerID).Order("id DESC").Limit(limit).Offset(offset).Find(&events).Error
	return events, err
}

func (r *eventRepository) CountByStatus(status models.EventStatus) (int64, error) {
	var count int64
	err := r.db.Model(&models.Event{}).Where("status = ?", status).Count(&count).Error
	return count, err
}
