// internal/repositories/payment_repository.go
package repositories

import (
	"eticketing/internal/models"
	"gorm.io/gorm"
)

type paymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) Create(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

func (r *paymentRepository) GetByID(id uint) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Preload("User").First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *paymentRepository) Update(payment *models.Payment) error {
	return r.db.Save(payment).Error
}

func (r *paymentRepository) ListByUser(userID uint, limit, offset int) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&payments).Error
	return payments, err
}

func (r *paymentRepository) GetTotalRevenue() (float64, error) {
	var total float64
	err := r.db.Model(&models.Payment{}).Where("status = ?", models.PaymentStatusCompleted).Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	return total, err
}

func (r *paymentRepository) CountTransactions() (int64, error) {
	var count int64
	err := r.db.Model(&models.Payment{}).Count(&count).Error
	return count, err
}
