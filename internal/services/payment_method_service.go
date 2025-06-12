// internal/services/payment_method_service.go
package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"

	"eticketing/internal/models"
	"eticketing/internal/repositories"
)

type PaymentMethodService struct {
	paymentMethodRepo repositories.PaymentMethodRepository
}

type CreatePaymentMethodRequest struct {
	UserID      uint                    `json:"-"` // Set by handler
	Type        models.PaymentType      `json:"type" binding:"required"`
	PaymentData CreatePaymentMethodData `json:"payment_data" binding:"required"`
	IsDefault   bool                    `json:"is_default"`
}

type CreatePaymentMethodData struct {
	// For Credit Card
	CardNumber string `json:"card_number,omitempty"`
	ExpiryDate string `json:"expiry_date,omitempty"`
	CVV        string `json:"cvv,omitempty"`
	CardHolder string `json:"card_holder,omitempty"`

	// For PayPal
	PayPalEmail string `json:"paypal_email,omitempty"`

	// For Apple Pay
	AppleID string `json:"apple_id,omitempty"`

	// For Google Pay
	GoogleEmail string `json:"google_email,omitempty"`
}

type UpdatePaymentMethodRequest struct {
	IsDefault *bool   `json:"is_default,omitempty"`
	Nickname  *string `json:"nickname,omitempty"`
}

type PaymentMethodResponse struct {
	ID         uint               `json:"id"`
	Type       models.PaymentType `json:"type"`
	TypeName   string             `json:"type_name"`
	MaskedData map[string]string  `json:"masked_data"`
	Token      string             `json:"token"`
	IsDefault  bool               `json:"is_default"`
	Nickname   string             `json:"nickname,omitempty"`
}

func NewPaymentMethodService(paymentMethodRepo repositories.PaymentMethodRepository) *PaymentMethodService {
	return &PaymentMethodService{
		paymentMethodRepo: paymentMethodRepo,
	}
}

func (s *PaymentMethodService) CreatePaymentMethod(req *CreatePaymentMethodRequest) (*PaymentMethodResponse, error) {
	// Validate payment data based on type
	if err := s.validatePaymentData(req.Type, req.PaymentData); err != nil {
		return nil, err
	}

	// Generate mock token
	token, err := s.generateMockToken()
	if err != nil {
		return nil, errors.New("failed to generate payment token")
	}

	// Convert payment data to JSON
	dataJSON, err := json.Marshal(req.PaymentData)
	if err != nil {
		return nil, errors.New("failed to process payment data")
	}

	// If this is the first payment method, make it default
	if req.IsDefault {
		err := s.paymentMethodRepo.ClearDefaultForUser(req.UserID)
		if err != nil {
			return nil, errors.New("failed to clear existing default")
		}
	}

	// Create payment method
	paymentMethod := &models.PaymentMethod{
		Type:      req.Type,
		Token:     token,
		Data:      string(dataJSON),
		UserID:    req.UserID,
		IsDefault: req.IsDefault,
	}

	if err := s.paymentMethodRepo.Create(paymentMethod); err != nil {
		return nil, errors.New("failed to create payment method")
	}

	return s.buildPaymentMethodResponse(paymentMethod), nil
}

func (s *PaymentMethodService) GetUserPaymentMethods(userID uint) ([]PaymentMethodResponse, error) {
	methods, err := s.paymentMethodRepo.ListByUser(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve payment methods")
	}

	var responses []PaymentMethodResponse
	for _, method := range methods {
		responses = append(responses, *s.buildPaymentMethodResponse(&method))
	}

	return responses, nil
}

func (s *PaymentMethodService) GetPaymentMethod(methodID, userID uint) (*PaymentMethodResponse, error) {
	method, err := s.paymentMethodRepo.GetByID(methodID)
	if err != nil {
		return nil, errors.New("payment method not found")
	}

	if method.UserID != userID {
		return nil, errors.New("unauthorized to access this payment method")
	}

	return s.buildPaymentMethodResponse(method), nil
}

func (s *PaymentMethodService) UpdatePaymentMethod(methodID, userID uint, req *UpdatePaymentMethodRequest) error {
	method, err := s.paymentMethodRepo.GetByID(methodID)
	if err != nil {
		return errors.New("payment method not found")
	}

	if method.UserID != userID {
		return errors.New("unauthorized to update this payment method")
	}

	if req.IsDefault != nil && *req.IsDefault {
		err := s.paymentMethodRepo.ClearDefaultForUser(userID)
		if err != nil {
			return errors.New("failed to clear existing default")
		}
		method.IsDefault = true
	}

	if err := s.paymentMethodRepo.Update(method); err != nil {
		return errors.New("failed to update payment method")
	}

	return nil
}

func (s *PaymentMethodService) DeletePaymentMethod(methodID, userID uint) error {
	method, err := s.paymentMethodRepo.GetByID(methodID)
	if err != nil {
		return errors.New("payment method not found")
	}

	if method.UserID != userID {
		return errors.New("unauthorized to delete this payment method")
	}

	if err := s.paymentMethodRepo.Delete(methodID); err != nil {
		return errors.New("failed to delete payment method")
	}

	return nil
}

func (s *PaymentMethodService) SetDefaultPaymentMethod(methodID, userID uint) error {
	method, err := s.paymentMethodRepo.GetByID(methodID)
	if err != nil {
		return errors.New("payment method not found")
	}

	if method.UserID != userID {
		return errors.New("unauthorized to modify this payment method")
	}

	// Clear existing default
	err = s.paymentMethodRepo.ClearDefaultForUser(userID)
	if err != nil {
		return errors.New("failed to clear existing default")
	}

	// Set new default
	method.IsDefault = true
	if err := s.paymentMethodRepo.Update(method); err != nil {
		return errors.New("failed to set default payment method")
	}

	return nil
}

func (s *PaymentMethodService) validatePaymentData(paymentType models.PaymentType, data CreatePaymentMethodData) error {
	switch paymentType {
	case models.PaymentTypeCard:
		if data.CardNumber == "" || data.ExpiryDate == "" || data.CVV == "" || data.CardHolder == "" {
			return errors.New("card number, expiry date, CVV, and card holder are required for credit card")
		}
	case models.PaymentTypePayPal:
		if data.PayPalEmail == "" {
			return errors.New("PayPal email is required")
		}
	case models.PaymentTypeGooglePay:
		if data.GoogleEmail == "" {
			return errors.New("Google email is required")
		}
	default:
		return errors.New("unsupported payment type")
	}
	return nil
}

func (s *PaymentMethodService) generateMockToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return "mock_" + hex.EncodeToString(bytes)[:32], nil
}

func (s *PaymentMethodService) buildPaymentMethodResponse(method *models.PaymentMethod) *PaymentMethodResponse {
	var data CreatePaymentMethodData
	_ = json.Unmarshal([]byte(method.Data), &data)

	response := &PaymentMethodResponse{
		ID:         method.ID,
		Type:       method.Type,
		TypeName:   s.getPaymentTypeName(method.Type),
		Token:      method.Token,
		IsDefault:  method.IsDefault,
		MaskedData: make(map[string]string),
	}

	// Create masked data based on type
	switch method.Type {
	case models.PaymentTypeCard:
		response.MaskedData["card_number"] = s.maskCardNumber(data.CardNumber)
		response.MaskedData["card_holder"] = data.CardHolder
		response.MaskedData["expiry_date"] = data.ExpiryDate
	case models.PaymentTypePayPal:
		response.MaskedData["email"] = s.maskEmail(data.PayPalEmail)
	case models.PaymentTypeGooglePay:
		response.MaskedData["email"] = s.maskEmail(data.GoogleEmail)
	}

	return response
}

func (s *PaymentMethodService) getPaymentTypeName(paymentType models.PaymentType) string {
	switch paymentType {
	case models.PaymentTypeCard:
		return "Credit Card"
	case models.PaymentTypePayPal:
		return "PayPal"
	case models.PaymentTypeGooglePay:
		return "Google Pay"
	case models.PaymentTypeStripe:
		return "Stripe"
	default:
		return "Unknown"
	}
}

func (s *PaymentMethodService) maskCardNumber(cardNumber string) string {
	if len(cardNumber) < 4 {
		return "****"
	}
	return "**** **** **** " + cardNumber[len(cardNumber)-4:]
}

func (s *PaymentMethodService) maskEmail(email string) string {
	if len(email) < 3 {
		return "***"
	}
	at := -1
	for i, c := range email {
		if c == '@' {
			at = i
			break
		}
	}
	if at == -1 {
		return "***"
	}

	prefix := email[:at]
	suffix := email[at:]

	if len(prefix) <= 2 {
		return "**" + suffix
	}

	return prefix[:2] + "***" + suffix
}
