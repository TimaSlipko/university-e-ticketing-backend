// internal/repositories/ticket_repository.go
package repositories

import (
	"eticketing/internal/models"
	"gorm.io/gorm"
)

type ticketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ticket *models.Ticket) error {
	return r.db.Create(ticket).Error
}

func (r *ticketRepository) GetByID(id uint) (*models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.Preload("Event").Preload("Sale").First(&ticket, id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) Update(ticket *models.Ticket) error {
	return r.db.Save(ticket).Error
}

func (r *ticketRepository) Delete(id uint) error {
	return r.db.Delete(&models.Ticket{}, id).Error
}

func (r *ticketRepository) ListByEvent(eventID uint) ([]models.Ticket, error) {
	var tickets []models.Ticket
	err := r.db.Where("event_id = ?", eventID).Find(&tickets).Error
	return tickets, err
}

func (r *ticketRepository) ListAvailableByEvent(eventID uint) ([]models.Ticket, error) {
	var tickets []models.Ticket
	err := r.db.Where("event_id = ? AND is_sold = false AND is_held = false", eventID).Find(&tickets).Error
	return tickets, err
}

func (r *ticketRepository) CountAvailableByEvent(eventID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Ticket{}).Where("event_id = ? AND is_sold = false AND is_held = false", eventID).Count(&count).Error
	return count, err
}
