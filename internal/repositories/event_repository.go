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

func (r *eventRepository) CountBySellerAndStatus(sellerID uint, status models.EventStatus) (int64, error) {
	var count int64
	query := r.db.Model(&models.Event{}).Where("seller_id = ?", sellerID)

	if status > 0 {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *eventRepository) CountEventsWithSoldTickets(sellerID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Event{}).
		Where("seller_id = ? AND id IN (SELECT DISTINCT event_id FROM tickets WHERE is_sold = true)", sellerID).
		Count(&count).Error
	return count, err
}

// Add to repositories/payment_repository.go
func (r *paymentRepository) GetTotalRevenueByUser(userID uint, userType models.UserType) (float64, error) {
	var total float64
	err := r.db.Model(&models.Payment{}).
		Where("user_id = ? AND user_type = ? AND status = ?", userID, userType, models.PaymentStatusCompleted).
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	return total, err
}

func (r *paymentRepository) GetPendingRevenueByUser(userID uint, userType models.UserType) (float64, error) {
	var total float64
	err := r.db.Model(&models.Payment{}).
		Where("user_id = ? AND user_type = ? AND status = ?", userID, userType, models.PaymentStatusPending).
		Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	return total, err
}

// Add to repositories/ticket_repository.go
type TicketStats struct {
	TotalTickets int64 `json:"total_tickets"`
	SoldTickets  int64 `json:"sold_tickets"`
}
