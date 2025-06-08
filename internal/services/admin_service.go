// internal/services/admin_service.go
package services

import (
	"errors"

	"eticketing/internal/models"
	"eticketing/internal/repositories"
	"eticketing/internal/utils"
	"gorm.io/gorm"
)

type AdminService struct {
	adminRepo   repositories.AdminRepository
	userRepo    repositories.UserRepository
	sellerRepo  repositories.SellerRepository
	eventRepo   repositories.EventRepository
	paymentRepo repositories.PaymentRepository
}

type AdminInfo struct {
	ID        uint            `json:"id"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
	Name      string          `json:"name"`
	Surname   string          `json:"surname"`
	UserType  models.UserType `json:"user_type"`
	AdminRole int             `json:"admin_role"`
}

type SystemStats struct {
	TotalUsers        int64   `json:"total_users"`
	TotalSellers      int64   `json:"total_sellers"`
	TotalAdmins       int64   `json:"total_admins"`
	PendingEvents     int64   `json:"pending_events"`
	ApprovedEvents    int64   `json:"approved_events"`
	TotalRevenue      float64 `json:"total_revenue"`
	TotalTransactions int64   `json:"total_transactions"`
}

type EventApprovalRequest struct {
	EventID uint   `json:"event_id" binding:"required"`
	Reason  string `json:"reason"`
}

func NewAdminService(
	adminRepo repositories.AdminRepository,
	userRepo repositories.UserRepository,
	sellerRepo repositories.SellerRepository,
	eventRepo repositories.EventRepository,
	paymentRepo repositories.PaymentRepository,
) *AdminService {
	return &AdminService{
		adminRepo:   adminRepo,
		userRepo:    userRepo,
		sellerRepo:  sellerRepo,
		eventRepo:   eventRepo,
		paymentRepo: paymentRepo,
	}
}

func (s *AdminService) GetProfile(adminID uint) (*AdminInfo, error) {
	admin, err := s.adminRepo.GetByID(adminID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("admin not found")
		}
		return nil, errors.New("failed to get admin profile")
	}

	return &AdminInfo{
		ID:        admin.ID,
		Username:  admin.Username,
		Email:     admin.Email,
		Name:      admin.Name,
		Surname:   admin.Surname,
		UserType:  models.UserTypeAdmin,
		AdminRole: admin.AdminRole,
	}, nil
}

func (s *AdminService) UpdateProfile(adminID uint, req *UpdateProfileRequest) (*AdminInfo, error) {
	admin, err := s.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, errors.New("admin not found")
	}

	// Validate username if provided
	if req.Username != "" && req.Username != admin.Username {
		if valid, validationErrors := utils.ValidateUsername(req.Username); !valid {
			return nil, errors.New("username validation failed: " + validationErrors[0])
		}

		// Check if username is already taken
		if existingAdmin, _ := s.adminRepo.GetByUsername(req.Username); existingAdmin != nil && existingAdmin.ID != adminID {
			return nil, errors.New("username already taken")
		}
		admin.Username = utils.SanitizeString(req.Username)
	}

	// Update other fields
	if req.Name != "" {
		admin.Name = utils.SanitizeString(req.Name)
	}
	if req.Surname != "" {
		admin.Surname = utils.SanitizeString(req.Surname)
	}

	if err := s.adminRepo.Update(admin); err != nil {
		return nil, errors.New("failed to update profile")
	}

	return &AdminInfo{
		ID:        admin.ID,
		Username:  admin.Username,
		Email:     admin.Email,
		Name:      admin.Name,
		Surname:   admin.Surname,
		UserType:  models.UserTypeAdmin,
		AdminRole: admin.AdminRole,
	}, nil
}

func (s *AdminService) ChangePassword(adminID uint, req *ChangePasswordRequest) error {
	admin, err := s.adminRepo.GetByID(adminID)
	if err != nil {
		return errors.New("admin not found")
	}

	// Verify current password
	if !utils.CheckPassword(req.CurrentPassword, admin.PasswordHash) {
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

	admin.PasswordHash = hashedPassword
	if err := s.adminRepo.Update(admin); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

func (s *AdminService) GetSystemStats() (*SystemStats, error) {
	// Get user counts
	totalUsers, err := s.userRepo.Count()
	if err != nil {
		return nil, errors.New("failed to count users")
	}

	totalSellers, err := s.sellerRepo.Count()
	if err != nil {
		return nil, errors.New("failed to count sellers")
	}

	totalAdmins, err := s.adminRepo.Count()
	if err != nil {
		return nil, errors.New("failed to count admins")
	}

	// Get event counts
	pendingEvents, err := s.eventRepo.CountByStatus(models.EventStatusPending)
	if err != nil {
		return nil, errors.New("failed to count pending events")
	}

	approvedEvents, err := s.eventRepo.CountByStatus(models.EventStatusApproved)
	if err != nil {
		return nil, errors.New("failed to count approved events")
	}

	// Get payment stats
	totalRevenue, err := s.paymentRepo.GetTotalRevenue()
	if err != nil {
		return nil, errors.New("failed to get total revenue")
	}

	totalTransactions, err := s.paymentRepo.CountTransactions()
	if err != nil {
		return nil, errors.New("failed to count transactions")
	}

	return &SystemStats{
		TotalUsers:        totalUsers,
		TotalSellers:      totalSellers,
		TotalAdmins:       totalAdmins,
		PendingEvents:     pendingEvents,
		ApprovedEvents:    approvedEvents,
		TotalRevenue:      totalRevenue,
		TotalTransactions: totalTransactions,
	}, nil
}

func (s *AdminService) GetPendingEvents(page, limit int) (*utils.PaginatedResponse, error) {
	offset := (page - 1) * limit
	events, err := s.eventRepo.ListByStatus(models.EventStatusPending, limit, offset)
	if err != nil {
		return nil, errors.New("failed to retrieve pending events")
	}

	total, err := s.eventRepo.CountByStatus(models.EventStatusPending)
	if err != nil {
		return nil, errors.New("failed to count pending events")
	}

	// Convert to response format
	var eventResponses []EventResponse
	for _, event := range events {
		response := EventResponse{
			ID:          event.ID,
			Title:       event.Title,
			Description: event.Description,
			Date:        event.Date,
			Address:     event.Address,
			Data:        event.Data,
			Status:      event.Status,
			SellerID:    event.SellerID,
			SellerName:  event.Seller.Name + " " + event.Seller.Surname,
		}
		eventResponses = append(eventResponses, response)
	}

	pagination := utils.CalculatePagination(page, limit, total)

	return &utils.PaginatedResponse{
		Success:    true,
		Message:    "Pending events retrieved successfully",
		Data:       eventResponses,
		Pagination: pagination,
	}, nil
}

func (s *AdminService) ApproveEvent(eventID uint) error {
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}

	if event.Status != models.EventStatusPending {
		return errors.New("only pending events can be approved")
	}

	event.Status = models.EventStatusApproved
	if err := s.eventRepo.Update(event); err != nil {
		return errors.New("failed to approve event")
	}

	return nil
}

func (s *AdminService) RejectEvent(eventID uint, reason string) error {
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}

	if event.Status != models.EventStatusPending {
		return errors.New("only pending events can be rejected")
	}

	event.Status = models.EventStatusRejected
	// TODO: Store rejection reason in event data or create separate table
	if err := s.eventRepo.Update(event); err != nil {
		return errors.New("failed to reject event")
	}

	return nil
}
