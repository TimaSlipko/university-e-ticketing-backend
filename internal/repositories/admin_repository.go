// internal/repositories/admin_repository.go
package repositories

import (
	"eticketing/internal/models"
	"gorm.io/gorm"
)

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) Create(admin *models.Admin) error {
	return r.db.Create(admin).Error
}

func (r *adminRepository) GetByID(id uint) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.First(&admin, id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetByEmail(email string) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.Where("email = ?", email).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) GetByUsername(username string) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.Where("username = ?", username).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *adminRepository) Update(admin *models.Admin) error {
	return r.db.Save(admin).Error
}

func (r *adminRepository) Delete(id uint) error {
	return r.db.Delete(&models.Admin{}, id).Error
}

func (r *adminRepository) List(limit, offset int) ([]models.Admin, error) {
	var admins []models.Admin
	err := r.db.Limit(limit).Offset(offset).Find(&admins).Error
	return admins, err
}

func (r *adminRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.Admin{}).Count(&count).Error
	return count, err
}
