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
	sellerRepo repositories.SellerRepository
	eventRepo  repositories.EventRepository
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

func NewSellerService(sellerRepo repositories.SellerRepository, eventRepo repositories.EventRepository) *SellerService {
	return &SellerService{
		sellerRepo: sellerRepo,
		eventRepo:  eventRepo,
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

func (s *SellerService) GetSellerStats(sellerID uint) (*SellerStats, error) {
	// Get total events by seller
	allEvents, err := s.eventRepo.ListBySeller(sellerID, 1000, 0)
	if err != nil {
		return nil, errors.New("failed to get seller events")
	}

	totalEvents := int64(len(allEvents))

	// Count by status
	var approvedEvents, pendingEvents int64
	for _, event := range allEvents {
		switch event.Status {
		case models.EventStatusApproved:
			approvedEvents++
		case models.EventStatusPending:
			pendingEvents++
		}
	}

	// TODO: Calculate actual revenue and events sold
	// This would require joining with ticket sales data

	return &SellerStats{
		TotalEvents:    totalEvents,
		ApprovedEvents: approvedEvents,
		PendingEvents:  pendingEvents,
		TotalRevenue:   0.0, // TODO: Calculate from actual sales
		EventsSold:     0,   // TODO: Calculate events with sold tickets
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
