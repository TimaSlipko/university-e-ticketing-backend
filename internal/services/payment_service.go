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
	mockMode    bool
}

type PaymentRequest struct {
	UserID        uint               `json:"user_id"`
	Amount        float64            `json:"amount"`
	PaymentMethod models.PaymentType `json:"payment_method"`
	Description   string             `json:"description"`
}

type PaymentResponse struct {
	PaymentID     uint                 `json:"payment_id"`
	Status        models.PaymentStatus `json:"status"`
	Amount        float64              `json:"amount"`
	TransactionID string               `json:"transaction_id"`
	Message       string               `json:"message"`
}

func NewPaymentService(paymentRepo repositories.PaymentRepository, mockMode bool) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		mockMode:    mockMode,
	}
}

func (s *PaymentService) ProcessPayment(req *PaymentRequest) (*PaymentResponse, error) {
	if req.Amount <= 0 {
		return nil, errors.New("payment amount must be greater than 0")
	}

	// Create payment record
	payment := &models.Payment{
		UserID: req.UserID,
		Date:   time.Now().Unix(),
		Type:   req.PaymentMethod,
		Amount: req.Amount,
		Status: models.PaymentStatusPending,
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, errors.New("failed to create payment record")
	}

	// Process payment (mocked)
	if s.mockMode {
		return s.processMockPayment(payment)
	}

	// In a real implementation, this would integrate with actual payment processors
	return nil, errors.New("real payment processing not implemented")
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
