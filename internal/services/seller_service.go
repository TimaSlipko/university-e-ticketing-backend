// internal/services/seller_service.go
package services

import (
	"errors"

	"eticketing/internal/models"
	"eticketing/internal/repositories"
	"eticketing/internal/utils"
	"gorm.io/gorm"
)

type SellerService struct {
	sellerRepo  repositories.SellerRepository
	eventRepo   repositories.EventRepository
	paymentRepo repositories.PaymentRepository // Add payment repo
	ticketRepo  repositories.TicketRepository  // Add ticket repo
}
type SellerInfo struct {
	ID       uint            `json:"id"`
	Username string          `json:"username"`
	Email    string          `json:"email"`
	Name     string          `json:"name"`
	Surname  string          `json:"surname"`
	UserType models.UserType `json:"user_type"`
}

type SellerStats struct {
	TotalEvents    int64   `json:"total_events"`
	ApprovedEvents int64   `json:"approved_events"`
	PendingEvents  int64   `json:"pending_events"`
	TotalRevenue   float64 `json:"total_revenue"`
	EventsSold     int64   `json:"events_sold"`
}

type SellerStatsResponse struct {
	TotalEvents    int     `json:"total_events"`
	ApprovedEvents int     `json:"approved_events"`
	PendingEvents  int     `json:"pending_events"`
	RejectedEvents int     `json:"rejected_events"`
	TotalRevenue   float64 `json:"total_revenue"`
	EventsSold     int     `json:"events_sold"`     // Events with sold tickets
	TotalTickets   int     `json:"total_tickets"`   // Total tickets created
	SoldTickets    int     `json:"sold_tickets"`    // Total tickets sold
	PendingRevenue float64 `json:"pending_revenue"` // Revenue from pending events
}

func NewSellerService(
	sellerRepo repositories.SellerRepository,
	eventRepo repositories.EventRepository,
	paymentRepo repositories.PaymentRepository,
	ticketRepo repositories.TicketRepository,
) *SellerService {
	return &SellerService{
		sellerRepo:  sellerRepo,
		eventRepo:   eventRepo,
		paymentRepo: paymentRepo,
		ticketRepo:  ticketRepo,
	}
}
func (s *SellerService) GetProfile(sellerID uint) (*SellerInfo, error) {
	seller, err := s.sellerRepo.GetByID(sellerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("seller not found")
		}
		return nil, errors.New("failed to get seller profile")
	}

	return &SellerInfo{
		ID:       seller.ID,
		Username: seller.Username,
		Email:    seller.Email,
		Name:     seller.Name,
		Surname:  seller.Surname,
		UserType: models.UserTypeSeller,
	}, nil
}

func (s *SellerService) UpdateProfile(sellerID uint, req *UpdateProfileRequest) (*SellerInfo, error) {
	seller, err := s.sellerRepo.GetByID(sellerID)
	if err != nil {
		return nil, errors.New("seller not found")
	}

	// Validate username if provided
	if req.Username != "" && req.Username != seller.Username {
		if valid, validationErrors := utils.ValidateUsername(req.Username); !valid {
			return nil, errors.New("username validation failed: " + validationErrors[0])
		}

		// Check if username is already taken
		if existingSeller, _ := s.sellerRepo.GetByUsername(req.Username); existingSeller != nil && existingSeller.ID != sellerID {
			return nil, errors.New("username already taken")
		}
		seller.Username = utils.SanitizeString(req.Username)
	}

	// Update other fields
	if req.Name != "" {
		seller.Name = utils.SanitizeString(req.Name)
	}
	if req.Surname != "" {
		seller.Surname = utils.SanitizeString(req.Surname)
	}

	if err := s.sellerRepo.Update(seller); err != nil {
		return nil, errors.New("failed to update profile")
	}

	return &SellerInfo{
		ID:       seller.ID,
		Username: seller.Username,
		Email:    seller.Email,
		Name:     seller.Name,
		Surname:  seller.Surname,
		UserType: models.UserTypeSeller,
	}, nil
}

func (s *SellerService) ChangePassword(sellerID uint, req *ChangePasswordRequest) error {
	seller, err := s.sellerRepo.GetByID(sellerID)
	if err != nil {
		return errors.New("seller not found")
	}

	// Verify current password
	if !utils.CheckPassword(req.CurrentPassword, seller.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	// Validate new password
	if valid, validationErrors := utils.ValidatePassword(req.NewPassword); !valid {
		return errors.New("password validation failed: " + validationErrors[0])
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	seller.PasswordHash = hashedPassword
	if err := s.sellerRepo.Update(seller); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (s *SellerService) GetSellerStats(sellerID uint) (*SellerStatsResponse, error) {
	// Get event counts by status
	totalEvents, err := s.eventRepo.CountBySellerAndStatus(sellerID, 0) // 0 = all statuses
	if err != nil {
		return nil, errors.New("failed to get total events count")
	}

	approvedEvents, err := s.eventRepo.CountBySellerAndStatus(sellerID, models.EventStatusApproved)
	if err != nil {
		return nil, errors.New("failed to get approved events count")
	}

	pendingEvents, err := s.eventRepo.CountBySellerAndStatus(sellerID, models.EventStatusPending)
	if err != nil {
		return nil, errors.New("failed to get pending events count")
	}

	rejectedEvents, err := s.eventRepo.CountBySellerAndStatus(sellerID, models.EventStatusRejected)
	if err != nil {
		return nil, errors.New("failed to get rejected events count")
	}

	// Get revenue from payments
	totalRevenue, err := s.paymentRepo.GetTotalRevenueByUser(sellerID, models.UserTypeSeller)
	if err != nil {
		return nil, errors.New("failed to get total revenue")
	}

	// Get ticket statistics
	ticketStats, err := s.ticketRepo.GetSellerTicketStats(sellerID)
	if err != nil {
		return nil, errors.New("failed to get ticket statistics")
	}

	// Count events with sold tickets
	eventsSold, err := s.eventRepo.CountEventsWithSoldTickets(sellerID)
	if err != nil {
		return nil, errors.New("failed to get events sold count")
	}

	// Get pending revenue (from pending events)
	pendingRevenue, err := s.paymentRepo.GetPendingRevenueByUser(sellerID, models.UserTypeSeller)
	if err != nil {
		pendingRevenue = 0 // Don't fail if this is not available
	}

	return &SellerStatsResponse{
		TotalEvents:    int(totalEvents),
		ApprovedEvents: int(approvedEvents),
		PendingEvents:  int(pendingEvents),
		RejectedEvents: int(rejectedEvents),
		TotalRevenue:   totalRevenue,
		EventsSold:     int(eventsSold),
		TotalTickets:   int(ticketStats.TotalTickets),
		SoldTickets:    int(ticketStats.SoldTickets),
		PendingRevenue: pendingRevenue,
	}, nil
}

func (s *SellerService) DeleteAccount(sellerID uint) error {
	// TODO: Add business logic to check if seller can be deleted
	// For example, check if they have active events, sold tickets, etc.

	seller, err := s.sellerRepo.GetByID(sellerID)
	if err != nil {
		return errors.New("seller not found")
	}

	if err := s.sellerRepo.Delete(seller.ID); err != nil {
		return errors.New("failed to delete account")
	}

	return nil
}
