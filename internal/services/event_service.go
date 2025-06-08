// internal/services/event_service.go
package services

import (
	"errors"
	"time"

	"eticketing/internal/models"
	"eticketing/internal/repositories"
	"eticketing/internal/utils"
)

type EventService struct {
	eventRepo  repositories.EventRepository
	ticketRepo repositories.TicketRepository
}

type CreateEventRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description" binding:"required"`
	Date        int64  `json:"date" binding:"required"`
	Address     string `json:"address" binding:"required"`
	Data        string `json:"data"`
	SellerID    uint   `json:"-"` // Set by handler
}

type UpdateEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Date        int64  `json:"date"`
	Address     string `json:"address"`
	Data        string `json:"data"`
}

type EventResponse struct {
	ID               uint               `json:"id"`
	Title            string             `json:"title"`
	Description      string             `json:"description"`
	Date             int64              `json:"date"`
	Address          string             `json:"address"`
	Data             string             `json:"data"`
	Status           models.EventStatus `json:"status"`
	SellerID         uint               `json:"seller_id"`
	SellerName       string             `json:"seller_name"`
	AvailableTickets int64              `json:"available_tickets"`
}

func NewEventService(eventRepo repositories.EventRepository, ticketRepo repositories.TicketRepository) *EventService {
	return &EventService{
		eventRepo:  eventRepo,
		ticketRepo: ticketRepo,
	}
}

func (s *EventService) CreateEvent(req *CreateEventRequest) (*EventResponse, error) {
	// Validate event date is in the future
	if req.Date <= time.Now().Unix() {
		return nil, errors.New("event date must be in the future")
	}

	event := &models.Event{
		Title:       utils.SanitizeString(req.Title),
		Description: utils.SanitizeString(req.Description),
		Date:        req.Date,
		Address:     utils.SanitizeString(req.Address),
		Data:        req.Data,
		SellerID:    req.SellerID,
		Status:      models.EventStatusPending,
	}

	if err := s.eventRepo.Create(event); err != nil {
		return nil, errors.New("failed to create event")
	}

	return s.eventToResponse(event), nil
}

func (s *EventService) GetEvents(page, limit int) (*utils.PaginatedResponse, error) {
	offset := (page - 1) * limit
	events, err := s.eventRepo.ListByStatus(models.EventStatusApproved, limit, offset)
	if err != nil {
		return nil, errors.New("failed to retrieve events")
	}

	total, err := s.eventRepo.CountByStatus(models.EventStatusApproved)
	if err != nil {
		return nil, errors.New("failed to count events")
	}

	var eventResponses []EventResponse
	for _, event := range events {
		availableTickets, _ := s.ticketRepo.CountAvailableByEvent(event.ID)
		response := s.eventToResponse(&event)
		response.AvailableTickets = availableTickets
		eventResponses = append(eventResponses, *response)
	}

	pagination := utils.CalculatePagination(page, limit, total)

	return &utils.PaginatedResponse{
		Success:    true,
		Message:    "Events retrieved successfully",
		Data:       eventResponses,
		Pagination: pagination,
	}, nil
}

func (s *EventService) GetEventsByStatus(status models.EventStatus, page, limit int) (*utils.PaginatedResponse, error) {
	offset := (page - 1) * limit
	events, err := s.eventRepo.ListByStatusReverse(status, limit, offset)
	if err != nil {
		return nil, errors.New("failed to retrieve events")
	}

	total, err := s.eventRepo.CountByStatus(status)
	if err != nil {
		return nil, errors.New("failed to count events")
	}

	var eventResponses []EventResponse
	for _, event := range events {
		availableTickets, _ := s.ticketRepo.CountAvailableByEvent(event.ID)
		response := s.eventToResponse(&event)
		response.AvailableTickets = availableTickets
		eventResponses = append(eventResponses, *response)
	}

	pagination := utils.CalculatePagination(page, limit, total)

	return &utils.PaginatedResponse{
		Success:    true,
		Message:    "Events retrieved successfully",
		Data:       eventResponses,
		Pagination: pagination,
	}, nil
}

func (s *EventService) GetEventsBySeller(sellerID uint, page, limit int) (*utils.PaginatedResponse, error) {
	offset := (page - 1) * limit
	events, err := s.eventRepo.ListBySeller(sellerID, limit, offset)
	if err != nil {
		return nil, errors.New("failed to retrieve seller events")
	}

	// Count total events for seller
	var total int64
	allEvents, _ := s.eventRepo.ListBySeller(sellerID, 1000, 0) // Get all for count
	total = int64(len(allEvents))

	var eventResponses []EventResponse
	for _, event := range events {
		availableTickets, _ := s.ticketRepo.CountAvailableByEvent(event.ID)
		response := s.eventToResponse(&event)
		response.AvailableTickets = availableTickets
		eventResponses = append(eventResponses, *response)
	}

	pagination := utils.CalculatePagination(page, limit, total)

	return &utils.PaginatedResponse{
		Success:    true,
		Message:    "Seller events retrieved successfully",
		Data:       eventResponses,
		Pagination: pagination,
	}, nil
}

func (s *EventService) GetEventByID(eventID uint) (*EventResponse, error) {
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	availableTickets, _ := s.ticketRepo.CountAvailableByEvent(event.ID)
	response := s.eventToResponse(event)
	response.AvailableTickets = availableTickets

	return response, nil
}

func (s *EventService) UpdateEvent(eventID, sellerID uint, req *UpdateEventRequest) (*EventResponse, error) {
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	// Check if seller owns the event
	if event.SellerID != sellerID {
		return nil, errors.New("unauthorized to update this event")
	}

	// Update fields if provided
	if req.Title != "" {
		event.Title = utils.SanitizeString(req.Title)
	}
	if req.Description != "" {
		event.Description = utils.SanitizeString(req.Description)
	}
	if req.Date != 0 {
		if req.Date <= time.Now().Unix() {
			return nil, errors.New("event date must be in the future")
		}
		event.Date = req.Date
	}
	if req.Address != "" {
		event.Address = utils.SanitizeString(req.Address)
	}
	if req.Data != "" {
		event.Data = req.Data
	}

	if err := s.eventRepo.Update(event); err != nil {
		return nil, errors.New("failed to update event")
	}

	return s.eventToResponse(event), nil
}

func (s *EventService) DeleteEvent(eventID, sellerID uint) error {
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}

	if event.SellerID != sellerID {
		return errors.New("unauthorized to delete this event")
	}

	// Check if event has sold tickets
	// TODO: Add logic to prevent deletion if tickets are sold

	if err := s.eventRepo.Delete(eventID); err != nil {
		return errors.New("failed to delete event")
	}

	return nil
}

func (s *EventService) eventToResponse(event *models.Event) *EventResponse {
	sellerName := ""
	if event.Seller.Name != "" {
		sellerName = event.Seller.Name + " " + event.Seller.Surname
	}

	return &EventResponse{
		ID:          event.ID,
		Title:       event.Title,
		Description: event.Description,
		Date:        event.Date,
		Address:     event.Address,
		Data:        event.Data,
		Status:      event.Status,
		SellerID:    event.SellerID,
		SellerName:  sellerName,
	}
}
