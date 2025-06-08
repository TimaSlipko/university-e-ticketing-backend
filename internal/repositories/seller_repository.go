// internal/repositories/seller_repository.go
package repositories

import (
	"eticketing/internal/models"
	"gorm.io/gorm"
)

type sellerRepository struct {
	db *gorm.DB
}

func NewSellerRepository(db *gorm.DB) SellerRepository {
	return &sellerRepository{db: db}
}

func (r *sellerRepository) Create(seller *models.Seller) error {
	return r.db.Create(seller).Error
}

func (r *sellerRepository) GetByID(id uint) (*models.Seller, error) {
	var seller models.Seller
	err := r.db.First(&seller, id).Error
	if err != nil {
		return nil, err
	}
	return &seller, nil
}

func (r *sellerRepository) GetByEmail(email string) (*models.Seller, error) {
	var seller models.Seller
	err := r.db.Where("email = ?", email).First(&seller).Error
	if err != nil {
		return nil, err
	}
	return &seller, nil
}

func (r *sellerRepository) GetByUsername(username string) (*models.Seller, error) {
	var seller models.Seller
	err := r.db.Where("username = ?", username).First(&seller).Error
	if err != nil {
		return nil, err
	}
	return &seller, nil
}

func (r *sellerRepository) Update(seller *models.Seller) error {
	return r.db.Save(seller).Error
}

func (r *sellerRepository) Delete(id uint) error {
	return r.db.Delete(&models.Seller{}, id).Error
}

func (r *sellerRepository) List(limit, offset int) ([]models.Seller, error) {
	var sellers []models.Seller
	err := r.db.Limit(limit).Offset(offset).Find(&sellers).Error
	return sellers, err
}

func (r *sellerRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Seller{}).Count(&count).Error
	return count, err
}
