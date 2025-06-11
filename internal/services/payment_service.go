// internal/services/payment_service.go
package services

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"eticketing/internal/models"
	"eticketing/internal/repositories"
)

type PaymentService struct {
	paymentRepo repositories.PaymentRepository
	eventRepo   repositories.EventRepository
	sellerRepo  repositories.SellerRepository
	mockMode    bool
}

type PaymentRequest struct {
	UserID        uint               `json:"user_id"`
	UserType      models.UserType    `json:"user_type"` // Add user type
	Amount        float64            `json:"amount"`
	PaymentMethod models.PaymentType `json:"payment_method"`
	Description   string             `json:"description"`
	EventID       uint               `json:"event_id,omitempty"`
}

type PaymentResponse struct {
	PaymentID     uint                 `json:"payment_id"`
	Status        models.PaymentStatus `json:"status"`
	Amount        float64              `json:"amount"`
	TransactionID string               `json:"transaction_id"`
	Message       string               `json:"message"`
}

type PaymentInfo struct {
	ID          uint                 `json:"id"`
	UserID      uint                 `json:"user_id"`
	Date        int64                `json:"date"`
	Type        models.PaymentType   `json:"type"`
	Amount      float64              `json:"amount"`
	Status      models.PaymentStatus `json:"status"`
	Description string               `json:"description"`
	EventTitle  string               `json:"event_title,omitempty"`
	PaymentType string               `json:"payment_type"` // "incoming" or "outgoing"
}

func NewPaymentService(paymentRepo repositories.PaymentRepository, eventRepo repositories.EventRepository, sellerRepo repositories.SellerRepository, mockMode bool) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		eventRepo:   eventRepo,
		sellerRepo:  sellerRepo,
		mockMode:    mockMode,
	}
}

func (s *PaymentService) ProcessPayment(req *PaymentRequest) (*PaymentResponse, error) {
	if req.Amount <= 0 {
		return nil, errors.New("payment amount must be greater than 0")
	}

	// Create customer payment record
	customerPayment := &models.Payment{
		UserID:      req.UserID,
		UserType:    req.UserType,
		Date:        time.Now().Unix(),
		Type:        req.PaymentMethod,
		Amount:      req.Amount,
		Status:      models.PaymentStatusPending,
		Description: req.Description,
		EventID:     req.EventID,
	}

	if err := s.paymentRepo.Create(customerPayment); err != nil {
		return nil, errors.New("failed to create payment record")
	}

	// Process payment (mocked)
	if s.mockMode {
		response, err := s.processMockPayment(customerPayment)
		if err != nil {
			return nil, err
		}

		// If payment successful and event_id provided, create seller payment
		if response.Status == models.PaymentStatusCompleted && req.EventID > 0 {
			err = s.createSellerPayment(req.EventID, req.Amount, req.Description)
			if err != nil {
				fmt.Printf("Failed to create seller payment: %v\n", err)
			}
		}

		return response, nil
	}

	return nil, errors.New("real payment processing not implemented")
}

func (s *PaymentService) createSellerPayment(eventID uint, amount float64, description string) error {
	// Get event to find seller
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return err
	}

	// Calculate seller fee (e.g., 95% to seller, 5% platform fee)
	sellerAmount := amount * 0.95

	// Create seller payment record
	sellerPayment := &models.Payment{
		UserID:      event.SellerID,
		UserType:    models.UserTypeSeller, // Set seller user type
		Date:        time.Now().Unix(),
		Type:        models.PaymentTypeCard,
		Amount:      sellerAmount,
		Status:      models.PaymentStatusCompleted,
		Description: fmt.Sprintf("Revenue from: %s", description),
		EventID:     eventID,
	}

	return s.paymentRepo.Create(sellerPayment)
}

func (s *PaymentService) GetUserPayments(userID uint, userType models.UserType, limit, offset int) ([]PaymentInfo, error) {
	payments, err := s.paymentRepo.ListByUserAndType(userID, userType, limit, offset)
	if err != nil {
		return nil, errors.New("failed to retrieve payments")
	}

	var paymentInfos []PaymentInfo
	for _, payment := range payments {
		paymentInfo := PaymentInfo{
			ID:          payment.ID,
			UserID:      payment.UserID,
			Date:        payment.Date,
			Type:        payment.Type,
			Amount:      payment.Amount,
			Status:      payment.Status,
			Description: payment.Description,
			PaymentType: s.getPaymentDirectionForUser(payment.UserType, userType),
		}

		// Add event title if available
		if payment.EventID > 0 {
			if event, err := s.eventRepo.GetByID(payment.EventID); err == nil {
				paymentInfo.EventTitle = event.Title
			}
		}

		paymentInfos = append(paymentInfos, paymentInfo)
	}

	return paymentInfos, nil
}

func (s *PaymentService) GetSellerPayments(sellerID uint, limit, offset int) ([]PaymentInfo, error) {
	payments, err := s.paymentRepo.ListByUser(sellerID, limit, offset)
	if err != nil {
		return nil, errors.New("failed to retrieve seller payments")
	}

	var paymentInfos []PaymentInfo
	for _, payment := range payments {
		paymentInfo := PaymentInfo{
			ID:          payment.ID,
			UserID:      payment.UserID,
			Date:        payment.Date,
			Type:        payment.Type,
			Amount:      payment.Amount,
			Status:      payment.Status,
			Description: payment.Description,
			PaymentType: "incoming", // Seller payments are incoming
		}

		// Add event title if available
		if payment.EventID > 0 {
			if event, err := s.eventRepo.GetByID(payment.EventID); err == nil {
				paymentInfo.EventTitle = event.Title
			}
		}

		paymentInfos = append(paymentInfos, paymentInfo)
	}

	return paymentInfos, nil
}

func (s *PaymentService) processMockPayment(payment *models.Payment) (*PaymentResponse, error) {
	// Simulate payment processing delay
	time.Sleep(time.Millisecond * 500)

	// Randomly succeed or fail (90% success rate for demo)
	rand.Seed(time.Now().UnixNano())
	success := rand.Float64() < 0.9

	if success {
		payment.Status = models.PaymentStatusCompleted
		transactionID := fmt.Sprintf("MOCK_%d_%d", payment.ID, time.Now().Unix())

		if err := s.paymentRepo.Update(payment); err != nil {
			return nil, errors.New("failed to update payment status")
		}

		return &PaymentResponse{
			PaymentID:     payment.ID,
			Status:        models.PaymentStatusCompleted,
			Amount:        payment.Amount,
			TransactionID: transactionID,
			Message:       "Payment processed successfully",
		}, nil
	} else {
		payment.Status = models.PaymentStatusFailed

		if err := s.paymentRepo.Update(payment); err != nil {
			return nil, errors.New("failed to update payment status")
		}

		return &PaymentResponse{
			PaymentID: payment.ID,
			Status:    models.PaymentStatusFailed,
			Amount:    payment.Amount,
			Message:   "Payment failed - insufficient funds or card declined",
		}, nil
	}
}

func (s *PaymentService) GetPaymentStatus(paymentID uint) (*PaymentResponse, error) {
	payment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return nil, errors.New("payment not found")
	}

	return &PaymentResponse{
		PaymentID: payment.ID,
		Status:    payment.Status,
		Amount:    payment.Amount,
		Message:   fmt.Sprintf("Payment is %d", payment.Status),
	}, nil
}

func (s *PaymentService) RefundPayment(paymentID uint) error {
	payment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return errors.New("payment not found")
	}

	if payment.Status != models.PaymentStatusCompleted {
		return errors.New("can only refund completed payments")
	}

	payment.Status = models.PaymentStatusRefunded
	if err := s.paymentRepo.Update(payment); err != nil {
		return errors.New("failed to process refund")
	}

	return nil
}

func (s *PaymentService) getPaymentDirectionForUser(paymentUserType, requestUserType models.UserType) string {
	if paymentUserType == models.UserTypeSeller && requestUserType == models.UserTypeSeller {
		return "incoming" // Seller viewing their revenue
	}
	return "outgoing" // User viewing their purchases
}
