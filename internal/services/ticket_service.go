// internal/services/ticket_service.go
package services

import (
	"errors"
	"eticketing/internal/models"
	"eticketing/internal/repositories"
	"time"
)

type TicketService struct {
	ticketRepo          repositories.TicketRepository
	purchasedTicketRepo repositories.PurchasedTicketRepository
	eventRepo           repositories.EventRepository
	saleRepo            repositories.SaleRepository
	paymentService      *PaymentService
}

type GroupedTicket = models.GroupedTicket

type CreateTicketRequest struct {
	Price       float64           `json:"price" binding:"required,min=0"`
	Type        models.TicketType `json:"type" binding:"required"`
	IsVip       bool              `json:"is_vip"`
	Title       string            `json:"title" binding:"required"`
	Description string            `json:"description"`
	Place       string            `json:"place" binding:"required"`
	SaleID      uint              `json:"sale_id" binding:"required"`
	EventID     uint              `json:"event_id" binding:"required"`
	Amount      int               `json:"amount" binding:"required,min=1,max=1000"`
}

type UpdateTicketRequest struct {
	Price       *float64           `json:"price"`
	Type        *models.TicketType `json:"type"`
	IsVip       *bool              `json:"is_vip"`
	Title       *string            `json:"title"`
	Description *string            `json:"description"`
	Place       *string            `json:"place"`
	SaleID      *uint              `json:"sale_id"`
}

type PurchaseTicketFromGroupRequest struct {
	UserID        uint               `json:"-"` // Set by handler
	EventID       uint               `json:"event_id" binding:"required"`
	Price         float64            `json:"price" binding:"required"`
	Type          models.TicketType  `json:"type" binding:"required"`
	IsVip         bool               `json:"is_vip"`
	Title         string             `json:"title" binding:"required"`
	Description   string             `json:"description"`
	Place         string             `json:"place" binding:"required"`
	SaleID        uint               `json:"sale_id" binding:"required"`
	Quantity      int                `json:"quantity" binding:"required,min=1,max=10"`
	PaymentMethod models.PaymentType `json:"payment_method" binding:"required"`
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
	EventID     uint    `json:"event_id"` // Add this field
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
	saleRepo repositories.SaleRepository,
	paymentService *PaymentService,
) *TicketService {
	return &TicketService{
		ticketRepo:          ticketRepo,
		purchasedTicketRepo: purchasedTicketRepo,
		eventRepo:           eventRepo,
		saleRepo:            saleRepo,
		paymentService:      paymentService,
	}
}

// New method for purchasing from grouped tickets with locking
func (s *TicketService) PurchaseTicketFromGroup(req *PurchaseTicketFromGroupRequest) (*PurchaseTicketResponse, error) {
	// Validate sale is active
	sale, err := s.saleRepo.GetByID(req.SaleID)
	if err != nil {
		return nil, errors.New("sale not found")
	}

	now := time.Now().Unix()
	if now < sale.StartDate || now > sale.EndDate {
		return nil, errors.New("sale is not currently active")
	}

	// Validate event
	event, err := s.eventRepo.GetByID(req.EventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	if event.Status != models.EventStatusApproved {
		return nil, errors.New("event is not approved for ticket sales")
	}

	// Begin transaction with locking
	availableTickets, err := s.ticketRepo.FindAndLockAvailableTickets(
		req.EventID, req.Price, req.Type, req.IsVip,
		req.Title, req.Place, req.SaleID, req.Quantity,
	)
	if err != nil {
		return nil, errors.New("failed to lock tickets: " + err.Error())
	}

	if len(availableTickets) < req.Quantity {
		return nil, errors.New("not enough tickets available")
	}

	// Calculate total amount
	totalAmount := req.Price * float64(req.Quantity)

	// Process payment
	paymentReq := &PaymentRequest{
		UserID:        req.UserID,
		Amount:        totalAmount,
		PaymentMethod: req.PaymentMethod,
		Description:   "Ticket purchase for " + req.Title + " - " + event.Title,
	}

	paymentResponse, err := s.paymentService.ProcessPayment(paymentReq)
	if err != nil {
		return nil, errors.New("payment processing failed: " + err.Error())
	}

	if paymentResponse.Status != models.PaymentStatusCompleted {
		return nil, errors.New("payment failed: " + paymentResponse.Message)
	}

	// Mark tickets as sold and create purchased ticket records
	var purchasedTickets []PurchasedTicketInfo
	for i := 0; i < req.Quantity; i++ {
		ticket := &availableTickets[i]

		// Mark as sold
		ticket.IsSold = true
		if err := s.ticketRepo.Update(ticket); err != nil {
			// TODO: Implement rollback mechanism
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
			// TODO: Implement rollback mechanism
			return nil, errors.New("failed to create purchased ticket record")
		}

		purchasedTickets = append(purchasedTickets, PurchasedTicketInfo{
			ID:          purchasedTicket.ID,
			TicketID:    ticket.ID,
			Title:       ticket.Title,
			Description: ticket.Description,
			Place:       ticket.Place,
			Price:       ticket.Price,
			EventTitle:  event.Title,
			EventDate:   event.Date,
			EventID:     event.ID, // Add this line
			IsUsed:      false,
		})
	}

	return &PurchaseTicketResponse{
		PurchasedTickets: purchasedTickets,
		PaymentInfo:      paymentResponse,
		TotalAmount:      totalAmount,
	}, nil
}

// Existing methods...

func (s *TicketService) CreateTickets(req *CreateTicketRequest, sellerID uint) error {
	// Verify event exists and belongs to seller
	event, err := s.eventRepo.GetByID(req.EventID)
	if err != nil {
		return errors.New("event not found")
	}
	if event.SellerID != sellerID {
		return errors.New("unauthorized to create tickets for this event")
	}

	// Verify sale exists and belongs to this event
	sale, err := s.saleRepo.GetByID(req.SaleID)
	if err != nil {
		return errors.New("sale not found")
	}
	if sale.EventID != req.EventID {
		return errors.New("sale does not belong to this event")
	}

	// Create the specified amount of tickets
	for i := 0; i < req.Amount; i++ {
		ticket := &models.Ticket{
			Price:       req.Price,
			Type:        req.Type,
			IsVip:       req.IsVip,
			Title:       req.Title,
			Description: req.Description,
			Place:       req.Place,
			SaleID:      req.SaleID,
			EventID:     req.EventID,
			IsSold:      false,
			IsHeld:      false,
		}

		if err := s.ticketRepo.Create(ticket); err != nil {
			return errors.New("failed to create tickets")
		}
	}

	return nil
}

func (s *TicketService) UpdateTickets(eventID uint, sellerID uint, oldTicket GroupedTicket, req *UpdateTicketRequest) error {
	// Verify event belongs to seller
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}
	if event.SellerID != sellerID {
		return errors.New("unauthorized to update tickets for this event")
	}

	// Find all tickets matching the old criteria (unsold only)
	tickets, err := s.ticketRepo.ListByGroupCriteria(eventID, oldTicket.Price, oldTicket.Type, oldTicket.IsVip, oldTicket.Title, oldTicket.Place, oldTicket.SaleID, false)
	if err != nil {
		return errors.New("failed to find tickets to update")
	}

	if len(tickets) == 0 {
		return errors.New("no unsold tickets found matching criteria")
	}

	// Update each ticket
	for _, ticket := range tickets {
		if req.Price != nil {
			ticket.Price = *req.Price
		}
		if req.Type != nil {
			ticket.Type = *req.Type
		}
		if req.IsVip != nil {
			ticket.IsVip = *req.IsVip
		}
		if req.Title != nil {
			ticket.Title = *req.Title
		}
		if req.Description != nil {
			ticket.Description = *req.Description
		}
		if req.Place != nil {
			ticket.Place = *req.Place
		}
		if req.SaleID != nil {
			// Verify new sale belongs to this event
			sale, err := s.saleRepo.GetByID(*req.SaleID)
			if err != nil {
				return errors.New("sale not found")
			}
			if sale.EventID != eventID {
				return errors.New("sale does not belong to this event")
			}
			ticket.SaleID = *req.SaleID
		}

		if err := s.ticketRepo.Update(&ticket); err != nil {
			return errors.New("failed to update tickets")
		}
	}

	return nil
}

func (s *TicketService) DeleteTickets(eventID uint, sellerID uint, groupedTicket GroupedTicket) error {
	// Verify event belongs to seller
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}
	if event.SellerID != sellerID {
		return errors.New("unauthorized to delete tickets for this event")
	}

	// Find all tickets matching the criteria (unsold only)
	tickets, err := s.ticketRepo.ListByGroupCriteria(eventID, groupedTicket.Price, groupedTicket.Type, groupedTicket.IsVip, groupedTicket.Title, groupedTicket.Place, groupedTicket.SaleID, false)
	if err != nil {
		return errors.New("failed to find tickets to delete")
	}

	if len(tickets) == 0 {
		return errors.New("no unsold tickets found matching criteria")
	}

	// Delete each ticket
	for _, ticket := range tickets {
		if err := s.ticketRepo.Delete(ticket.ID); err != nil {
			return errors.New("failed to delete tickets")
		}
	}

	return nil
}

func (s *TicketService) GetGroupedTicketsByEvent(eventID uint) ([]GroupedTicket, error) {
	groupedTickets, err := s.ticketRepo.ListGroupedByEvent(eventID)
	if err != nil {
		return nil, errors.New("failed to retrieve grouped tickets")
	}

	return groupedTickets, nil
}

func (s *TicketService) GetAvailableGroupedTicketsByEvent(eventID uint) ([]GroupedTicket, error) {
	groupedTickets, err := s.ticketRepo.ListAvailableGroupedByEvent(eventID)
	if err != nil {
		return nil, errors.New("failed to retrieve available grouped tickets")
	}

	return groupedTickets, nil
}

// Legacy methods for backward compatibility

func (s *TicketService) PurchaseTicket(req *PurchaseTicketRequest) (*PurchaseTicketResponse, error) {
	// Get ticket information with locking
	ticket, err := s.ticketRepo.GetByIDForUpdate(req.TicketID)
	if err != nil {
		return nil, errors.New("ticket not found")
	}

	if ticket.IsSold || ticket.IsHeld {
		return nil, errors.New("ticket is not available")
	}

	// Check if sale is active
	sale, err := s.saleRepo.GetByID(ticket.SaleID)
	if err != nil {
		return nil, errors.New("sale not found")
	}

	now := time.Now().Unix()
	if now < sale.StartDate || now > sale.EndDate {
		return nil, errors.New("sale is not currently active")
	}

	// Check if enough tickets are available (for quantity > 1, we'd need to implement bulk purchase)
	if req.Quantity > 1 {
		return nil, errors.New("bulk purchase not implemented for individual tickets")
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
				EventID:     ticket.EventID, // Add this line
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
			EventID:     ticket.Ticket.EventID, // Add this line
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
