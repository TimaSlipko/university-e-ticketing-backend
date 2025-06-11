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

func (r *ticketRepository) GetByIDForUpdate(id uint) (*models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.Preload("Event").Preload("Sale").
		Set("gorm:query_option", "FOR UPDATE").
		First(&ticket, id).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) FindAndLockAvailableTickets(
	eventID uint,
	price float64,
	ticketType models.TicketType,
	isVip bool,
	title, place string,
	saleID uint,
	quantity int,
) ([]models.Ticket, error) {
	var tickets []models.Ticket

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Use raw SQL with FOR UPDATE to lock the rows
		err := tx.
			Where("event_id = ? AND price = ? AND type = ? AND is_vip = ? AND title = ? AND place = ? AND sale_id = ? AND is_sold = false AND is_held = false",
				eventID, price, ticketType, isVip, title, place, saleID).
			Set("gorm:query_option", "FOR UPDATE").
			Limit(quantity).
			Find(&tickets).Error

		return err
	})

	return tickets, err
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

func (r *ticketRepository) ListByGroupCriteria(eventID uint, price float64, ticketType models.TicketType, isVip bool, title, place string, saleID uint, includeSold bool) ([]models.Ticket, error) {
	var tickets []models.Ticket
	query := r.db.Where("event_id = ? AND price = ? AND type = ? AND is_vip = ? AND title = ? AND place = ? AND sale_id = ?",
		eventID, price, ticketType, isVip, title, place, saleID)

	if !includeSold {
		query = query.Where("is_sold = false")
	}

	err := query.Find(&tickets).Error
	return tickets, err
}

func (r *ticketRepository) ListGroupedByEvent(eventID uint) ([]models.GroupedTicket, error) {
	var results []models.GroupedTicket

	err := r.db.Model(&models.Ticket{}).
		Select(`
			price, 
			type, 
			is_vip, 
			title, 
			description, 
			place, 
			sale_id, 
			event_id,
			COUNT(*) as total_amount,
			COUNT(CASE WHEN is_sold = false AND is_held = false THEN 1 END) as available_amount,
			COUNT(CASE WHEN is_sold = true THEN 1 END) as sold_amount,
			COUNT(CASE WHEN is_held = true AND is_sold = false THEN 1 END) as held_amount
		`).
		Where("event_id = ?", eventID).
		Group("price, type, is_vip, title, description, place, sale_id, event_id").
		Scan(&results).Error

	return results, err
}

func (r *ticketRepository) ListAvailableGroupedByEvent(eventID uint) ([]models.GroupedTicket, error) {
	var results []models.GroupedTicket

	err := r.db.Model(&models.Ticket{}).
		Select(`
			price, 
			type, 
			is_vip, 
			title, 
			description, 
			place, 
			sale_id, 
			event_id,
			COUNT(*) as total_amount,
			COUNT(CASE WHEN is_sold = false AND is_held = false THEN 1 END) as available_amount,
			COUNT(CASE WHEN is_sold = true THEN 1 END) as sold_amount,
			COUNT(CASE WHEN is_held = true AND is_sold = false THEN 1 END) as held_amount
		`).
		Where("event_id = ?", eventID).
		Group("price, type, is_vip, title, description, place, sale_id, event_id").
		Having("COUNT(CASE WHEN is_sold = false AND is_held = false THEN 1 END) > 0").
		Scan(&results).Error

	return results, err
}

func (r *ticketRepository) GetSellerTicketStats(sellerID uint) (*TicketStats, error) {
	var stats TicketStats

	// Get total tickets for seller's events
	err := r.db.Model(&models.Ticket{}).
		Joins("JOIN events ON tickets.event_id = events.id").
		Where("events.seller_id = ?", sellerID).
		Count(&stats.TotalTickets).Error // Fix: Remove the slice syntax
	if err != nil {
		return nil, err
	}

	// Get sold tickets for seller's events
	err = r.db.Model(&models.Ticket{}).
		Joins("JOIN events ON tickets.event_id = events.id").
		Where("events.seller_id = ? AND tickets.is_sold = true", sellerID).
		Count(&stats.SoldTickets).Error // Fix: Remove the slice syntax
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
