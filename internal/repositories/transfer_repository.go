// internal/repositories/transfer_repository.go
package repositories

import (
	"eticketing/internal/models"
	"gorm.io/gorm"
)

type transferRepository struct {
	db *gorm.DB
}

func NewTransferRepository(db *gorm.DB) TransferRepository {
	return &transferRepository{db: db}
}

func (r *transferRepository) CreateActive(transfer *models.ActiveTicketTransfer) error {
	return r.db.Create(transfer).Error
}

func (r *transferRepository) GetActiveByID(id uint) (*models.ActiveTicketTransfer, error) {
	var transfer models.ActiveTicketTransfer
	err := r.db.Preload("FromUser").Preload("ToUser").Preload("PurchasedTicket").First(&transfer, id).Error
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (r *transferRepository) UpdateActive(transfer *models.ActiveTicketTransfer) error {
	return r.db.Save(transfer).Error
}

func (r *transferRepository) CreateDone(transfer *models.DoneTicketTransfer) error {
	return r.db.Create(transfer).Error
}

func (r *transferRepository) ListActiveByUser(userID uint) ([]models.ActiveTicketTransfer, error) {
	var transfers []models.ActiveTicketTransfer
	err := r.db.Preload("FromUser").Preload("ToUser").Preload("PurchasedTicket").
		Where("(from_user_id = ? OR to_user_id = ?) AND status = ?", userID, userID, models.TransferStatusPending).
		Find(&transfers).Error
	return transfers, err
}

func (r *transferRepository) ListDoneByUser(userID uint) ([]models.DoneTicketTransfer, error) {
	var transfers []models.DoneTicketTransfer
	err := r.db.Preload("FromUser").Preload("ToUser").Preload("PurchasedTicket").
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).Find(&transfers).Error
	return transfers, err
}

func (r *transferRepository) HasActiveTransferForTicket(ticketID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.ActiveTicketTransfer{}).
		Where("purchased_ticket_id = ? AND status = ?", ticketID, models.TransferStatusPending).
		Count(&count).Error
	return count > 0, err
}

func (r *transferRepository) ListRejectedByUser(userID uint) ([]models.ActiveTicketTransfer, error) {
	var transfers []models.ActiveTicketTransfer
	err := r.db.Preload("FromUser").Preload("ToUser").Preload("PurchasedTicket").
		Where("(from_user_id = ? OR to_user_id = ?) AND (status = ? OR status = ?)",
			userID, userID, models.TransferStatusRejected, models.TransferStatusCancelled).
		Find(&transfers).Error
	return transfers, err
}
