package repositories

import (
	"eticketing/internal/models"
	"gorm.io/gorm"
)

type saleRepository struct {
	db *gorm.DB
}

func NewSaleRepository(db *gorm.DB) SaleRepository {
	return &saleRepository{db: db}
}

func (r *saleRepository) Create(sale *models.Sale) error {
	return r.db.Create(sale).Error
}

func (r *saleRepository) GetByID(id uint) (*models.Sale, error) {
	var sale models.Sale
	err := r.db.Preload("Event").First(&sale, id).Error
	if err != nil {
		return nil, err
	}
	return &sale, nil
}

func (r *saleRepository) Update(sale *models.Sale) error {
	return r.db.Save(sale).Error
}

func (r *saleRepository) Delete(id uint) error {
	return r.db.Delete(&models.Sale{}, id).Error
}

func (r *saleRepository) ListByEvent(eventID uint) ([]models.Sale, error) {
	var sales []models.Sale
	err := r.db.Where("event_id = ?", eventID).Order("start_date").Find(&sales).Error
	return sales, err
}
