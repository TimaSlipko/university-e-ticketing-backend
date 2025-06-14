// internal/repositories/purchased_ticket_repository.go
package repositories

import (
	"errors"
	"eticketing/internal/models"
	"gorm.io/gorm"
)

type purchasedTicketRepository struct {
	db *gorm.DB
}

func NewPurchasedTicketRepository(db *gorm.DB) PurchasedTicketRepository {
	return &purchasedTicketRepository{db: db}
}

func (r *purchasedTicketRepository) Create(ticket *models.PurchasedTicket) error {
	return r.db.Create(ticket).Error
}

func (r *purchasedTicketRepository) GetByID(id uint) (*models.PurchasedTicket, error) {
	var ticket models.PurchasedTicket
	err := r.db.Preload("User").Preload("Ticket").Preload("Ticket.Event").First(&ticket, id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *purchasedTicketRepository) ListByUser(userID uint) ([]models.PurchasedTicket, error) {
	var tickets []models.PurchasedTicket
	err := r.db.Preload("Ticket").Preload("Ticket.Event").Where("user_id = ?", userID).Find(&tickets).Error
	return tickets, err
}

func (r *purchasedTicketRepository) CountByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.PurchasedTicket{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *purchasedTicketRepository) UpdateOwnership(ticketID uint, newUserID uint) error {
	result := r.db.Exec("UPDATE purchased_tickets SET user_id = ? WHERE id = ?", newUserID, ticketID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no rows updated")
	}
	return nil
}
