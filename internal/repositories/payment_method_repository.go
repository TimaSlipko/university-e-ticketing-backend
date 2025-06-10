// internal/repositories/payment_method_repository.go
package repositories

import (
	"eticketing/internal/models"
	"gorm.io/gorm"
)

type paymentMethodRepository struct {
	db *gorm.DB
}

func NewPaymentMethodRepository(db *gorm.DB) PaymentMethodRepository {
	return &paymentMethodRepository{db: db}
}

func (r *paymentMethodRepository) Create(method *models.PaymentMethod) error {
	return r.db.Create(method).Error
}

func (r *paymentMethodRepository) GetByID(id uint) (*models.PaymentMethod, error) {
	var method models.PaymentMethod
	err := r.db.First(&method, id).Error
	if err != nil {
		return nil, err
	}
	return &method, nil
}

func (r *paymentMethodRepository) Update(method *models.PaymentMethod) error {
	return r.db.Save(method).Error
}

func (r *paymentMethodRepository) Delete(id uint) error {
	return r.db.Delete(&models.PaymentMethod{}, id).Error
}

func (r *paymentMethodRepository) ListByUser(userID uint) ([]models.PaymentMethod, error) {
	var methods []models.PaymentMethod
	err := r.db.Where("user_id = ?", userID).Order("is_default DESC, id ASC").Find(&methods).Error
	return methods, err
}

func (r *paymentMethodRepository) ClearDefaultForUser(userID uint) error {
	return r.db.Model(&models.PaymentMethod{}).
		Where("user_id = ? AND is_default = true", userID).
		Update("is_default", false).Error
}

func (r *paymentMethodRepository) GetDefaultByUser(userID uint) (*models.PaymentMethod, error) {
	var method models.PaymentMethod
	err := r.db.Where("user_id = ? AND is_default = true", userID).First(&method).Error
	if err != nil {
		return nil, err
	}
	return &method, nil
}
