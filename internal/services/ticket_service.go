// internal/services/ticket_service.go
package services

import (
	"errors"
	"eticketing/internal/models"
	"eticketing/internal/repositories"
)

type TicketService struct {
	ticketRepo          repositories.TicketRepository
	purchasedTicketRepo repositories.PurchasedTicketRepository
	eventRepo           repositories.EventRepository
	paymentService      *PaymentService
}

type PurchaseTicketRequest struct {
	UserID        uint               `json:"-"` // Set by handler
	TicketID      uint               `json:"ticket_id" binding:"required"`
	Quantity      int                `json:"quantity" binding:"required,min=1,max=10"`
	PaymentMethod models.PaymentType `json:"payment_method" binding:"required"`
}

type PurchaseTicketResponse struct {
	PurchasedTickets []PurchasedTicketInfo `json:"purchased_tickets"`
	PaymentInfo      *PaymentResponse      `json:"payment_info"`
	TotalAmount      float64               `json:"total_amount"`
}

type PurchasedTicketInfo struct {
	ID          uint    `json:"id"`
	TicketID    uint    `json:"ticket_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Place       string  `json:"place"`
	Price       float64 `json:"price"`
	EventTitle  string  `json:"event_title"`
	EventDate   int64   `json:"event_date"`
	IsUsed      bool    `json:"is_used"`
}

type TransferTicketRequest struct {
	FromUserID        uint   `json:"-"` // Set by handler
	ToUserEmail       string `json:"to_user_email" binding:"required,email"`
	PurchasedTicketID uint   `json:"purchased_ticket_id" binding:"required"`
}

func NewTicketService(
	ticketRepo repositories.TicketRepository,
	purchasedTicketRepo repositories.PurchasedTicketRepository,
	eventRepo repositories.EventRepository,
	paymentService *PaymentService,
) *TicketService {
	return &TicketService{
		ticketRepo:          ticketRepo,
		purchasedTicketRepo: purchasedTicketRepo,
		eventRepo:           eventRepo,
		paymentService:      paymentService,
	}
}

func (s *TicketService) PurchaseTicket(req *PurchaseTicketRequest) (*PurchaseTicketResponse, error) {
	// Get ticket information
	ticket, err := s.ticketRepo.GetByID(req.TicketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	if ticket.IsSold || ticket.IsHeld {
		return nil, errors.New("ticket is not available")
	}

	// Check if enough tickets are available (for quantity > 1, we'd need to implement bulk purchase)
	if req.Quantity > 1 {
		return nil, errors.New("bulk purchase not implemented yet")
	}

	// Calculate total amount
	totalAmount := ticket.Price * float64(req.Quantity)

	// Process payment
	paymentReq := &PaymentRequest{
		UserID:        req.UserID,
		Amount:        totalAmount,
		PaymentMethod: req.PaymentMethod,
		Description:   "Ticket purchase for " + ticket.Title,
	}

	paymentResponse, err := s.paymentService.ProcessPayment(paymentReq)
	if err != nil {
		return nil, errors.New("payment processing failed: " + err.Error())
	}

	if paymentResponse.Status != models.PaymentStatusCompleted {
		return nil, errors.New("payment failed: " + paymentResponse.Message)
	}

	// Mark ticket as sold
	ticket.IsSold = true
	if err := s.ticketRepo.Update(ticket); err != nil {
		// TODO: Refund payment here
		return nil, errors.New("failed to update ticket status")
	}

	// Create purchased ticket record
	purchasedTicket := &models.PurchasedTicket{
		Price:       ticket.Price,
		Type:        ticket.Type,
		IsVip:       ticket.IsVip,
		Title:       ticket.Title,
		Description: ticket.Description,
		Place:       ticket.Place,
		UserID:      req.UserID,
		TicketID:    ticket.ID,
	}

	if err := s.purchasedTicketRepo.Create(purchasedTicket); err != nil {
		return nil, errors.New("failed to create purchased ticket record")
	}

	// Get event info for response
	event, _ := s.eventRepo.GetByID(ticket.EventID)
	eventTitle := ""
	eventDate := int64(0)
	if event != nil {
		eventTitle = event.Title
		eventDate = event.Date
	}

	return &PurchaseTicketResponse{
		PurchasedTickets: []PurchasedTicketInfo{
			{
				ID:          purchasedTicket.ID,
				TicketID:    ticket.ID,
				Title:       ticket.Title,
				Description: ticket.Description,
				Place:       ticket.Place,
				Price:       ticket.Price,
				EventTitle:  eventTitle,
				EventDate:   eventDate,
				IsUsed:      false,
			},
		},
		PaymentInfo: paymentResponse,
		TotalAmount: totalAmount,
	}, nil
}

func (s *TicketService) GetUserTickets(userID uint) ([]PurchasedTicketInfo, error) {
	tickets, err := s.purchasedTicketRepo.ListByUser(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve user tickets")
	}

	var ticketInfos []PurchasedTicketInfo
	for _, ticket := range tickets {
		eventTitle := ""
		eventDate := int64(0)
		if ticket.Ticket.Event.Title != "" {
			eventTitle = ticket.Ticket.Event.Title
			eventDate = ticket.Ticket.Event.Date
		}

		ticketInfos = append(ticketInfos, PurchasedTicketInfo{
			ID:          ticket.ID,
			TicketID:    ticket.TicketID,
			Title:       ticket.Title,
			Description: ticket.Description,
			Place:       ticket.Place,
			Price:       ticket.Price,
			EventTitle:  eventTitle,
			EventDate:   eventDate,
			IsUsed:      ticket.IsUsed,
		})
	}

	return ticketInfos, nil
}

func (s *TicketService) GetEventTickets(eventID uint) ([]models.Ticket, error) {
	tickets, err := s.ticketRepo.ListAvailableByEvent(eventID)
	if err != nil {
		return nil, errors.New("failed to retrieve event tickets")
	}

	return tickets, nil
}

func (s *TicketService) TransferTicket(req *TransferTicketRequest) error {
	// Get purchased ticket
	purchasedTicket, err := s.purchasedTicketRepo.GetByID(req.PurchasedTicketID)
	if err != nil {
		return errors.New("purchased ticket not found")
	}

	// Check if user owns the ticket
	if purchasedTicket.UserID != req.FromUserID {
		return errors.New("unauthorized to transfer this ticket")
	}

	if purchasedTicket.IsUsed {
		return errors.New("cannot transfer used ticket")
	}

	// Note: This is a simplified implementation
	// The full transfer logic is now handled by TransferService
	// This method could be deprecated in favor of TransferService.InitiateTransfer

	return errors.New("use /api/v1/transfers endpoints for ticket transfers")
}
