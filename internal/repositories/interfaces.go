// internal/repositories/interfaces.go
package repositories

import (
	"eticketing/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	List(limit, offset int) ([]models.User, error)
	Count() (int64, error)
}

type SellerRepository interface {
	Create(seller *models.Seller) error
	GetByID(id uint) (*models.Seller, error)
	GetByEmail(email string) (*models.Seller, error)
	GetByUsername(username string) (*models.Seller, error)
	Update(seller *models.Seller) error
	Delete(id uint) error
	List(limit, offset int) ([]models.Seller, error)
	Count() (int64, error)
}

type AdminRepository interface {
	Create(admin *models.Admin) error
	GetByID(id uint) (*models.Admin, error)
	GetByEmail(email string) (*models.Admin, error)
	GetByUsername(username string) (*models.Admin, error)
	Update(admin *models.Admin) error
	Delete(id uint) error
	List(limit, offset int) ([]models.Admin, error)
	Count() (int64, error)
}

type EventRepository interface {
	Create(event *models.Event) error
	GetByID(id uint) (*models.Event, error)
	Update(event *models.Event) error
	Delete(id uint) error
	ListByStatus(status models.EventStatus, limit, offset int) ([]models.Event, error)
	ListByStatusReverse(status models.EventStatus, limit, offset int) ([]models.Event, error)
	ListBySeller(sellerID uint, limit, offset int) ([]models.Event, error)
	CountByStatus(status models.EventStatus) (int64, error)
}

type TicketRepository interface {
	Create(ticket *models.Ticket) error
	GetByID(id uint) (*models.Ticket, error)
	Update(ticket *models.Ticket) error
	Delete(id uint) error
	ListByEvent(eventID uint) ([]models.Ticket, error)
	ListAvailableByEvent(eventID uint) ([]models.Ticket, error)
	CountAvailableByEvent(eventID uint) (int64, error)
}

type PurchasedTicketRepository interface {
	Create(ticket *models.PurchasedTicket) error
	GetByID(id uint) (*models.PurchasedTicket, error)
	Update(ticket *models.PurchasedTicket) error
	ListByUser(userID uint) ([]models.PurchasedTicket, error)
	CountByUser(userID uint) (int64, error)
}

type PaymentRepository interface {
	Create(payment *models.Payment) error
	GetByID(id uint) (*models.Payment, error)
	Update(payment *models.Payment) error
	ListByUser(userID uint, limit, offset int) ([]models.Payment, error)
	GetTotalRevenue() (float64, error)
	CountTransactions() (int64, error)
}

type TransferRepository interface {
	CreateActive(transfer *models.ActiveTicketTransfer) error
	GetActiveByID(id uint) (*models.ActiveTicketTransfer, error)
	UpdateActive(transfer *models.ActiveTicketTransfer) error
	CreateDone(transfer *models.DoneTicketTransfer) error
	ListActiveByUser(userID uint) ([]models.ActiveTicketTransfer, error)
	ListDoneByUser(userID uint) ([]models.DoneTicketTransfer, error)
}
