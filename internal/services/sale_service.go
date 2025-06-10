package services

import (
	"errors"
	"time"

	"eticketing/internal/models"
	"eticketing/internal/repositories"
)

type SaleService struct {
	saleRepo  repositories.SaleRepository
	eventRepo repositories.EventRepository
}

type CreateSaleRequest struct {
	StartDate int64 `json:"start_date" binding:"required"`
	EndDate   int64 `json:"end_date" binding:"required"`
	EventID   uint  `json:"event_id" binding:"required"`
}

type UpdateSaleRequest struct {
	StartDate int64 `json:"start_date"`
	EndDate   int64 `json:"end_date"`
}

type SaleResponse struct {
	ID        uint  `json:"id"`
	StartDate int64 `json:"start_date"`
	EndDate   int64 `json:"end_date"`
	EventID   uint  `json:"event_id"`
	IsActive  bool  `json:"is_active"`
	EventInfo struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Date        int64  `json:"date"`
		Address     string `json:"address"`
	} `json:"event_info,omitempty"`
}

func NewSaleService(saleRepo repositories.SaleRepository, eventRepo repositories.EventRepository) *SaleService {
	return &SaleService{
		saleRepo:  saleRepo,
		eventRepo: eventRepo,
	}
}

func (s *SaleService) CreateSale(req *CreateSaleRequest, sellerID uint) (*SaleResponse, error) {
	// Validate dates
	now := time.Now().Unix()
	if req.StartDate <= now {
		return nil, errors.New("sale start date must be in the future")
	}
	if req.EndDate <= req.StartDate {
		return nil, errors.New("sale end date must be after start date")
	}

	// Verify event exists and belongs to seller
	event, err := s.eventRepo.GetByID(req.EventID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	if event.SellerID != sellerID {
		return nil, errors.New("unauthorized to create sale for this event")
	}

	// Check if event is approved
	if event.Status != models.EventStatusApproved {
		return nil, errors.New("can only create sales for approved events")
	}

	// Check for overlapping sales
	existingSales, err := s.saleRepo.ListByEvent(req.EventID)
	if err != nil {
		return nil, errors.New("failed to check existing sales")
	}

	for _, existingSale := range existingSales {
		if s.datesOverlap(req.StartDate, req.EndDate, existingSale.StartDate, existingSale.EndDate) {
			return nil, errors.New("sale dates overlap with existing sale")
		}
	}

	sale := &models.Sale{
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		EventID:   req.EventID,
	}

	if err := s.saleRepo.Create(sale); err != nil {
		return nil, errors.New("failed to create sale")
	}

	return s.saleToResponse(sale, event), nil
}

func (s *SaleService) GetSalesByEvent(eventID uint) ([]SaleResponse, error) {
	// Verify event exists
	event, err := s.eventRepo.GetByID(eventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	sales, err := s.saleRepo.ListByEvent(eventID)
	if err != nil {
		return nil, errors.New("failed to retrieve sales")
	}

	var saleResponses []SaleResponse
	for _, sale := range sales {
		response := s.saleToResponse(&sale, event)
		saleResponses = append(saleResponses, *response)
	}

	return saleResponses, nil
}

func (s *SaleService) GetSaleByID(saleID uint) (*SaleResponse, error) {
	sale, err := s.saleRepo.GetByID(saleID)
	if err != nil {
		return nil, errors.New("sale not found")
	}

	event, err := s.eventRepo.GetByID(sale.EventID)
	if err != nil {
		return nil, errors.New("event not found")
	}

	return s.saleToResponse(sale, event), nil
}

func (s *SaleService) UpdateSale(saleID, sellerID uint, req *UpdateSaleRequest) (*SaleResponse, error) {
	sale, err := s.saleRepo.GetByID(saleID)
	if err != nil {
		return nil, errors.New("sale not found")
	}

	// Verify seller owns the event
	event, err := s.eventRepo.GetByID(sale.EventID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	if event.SellerID != sellerID {
		return nil, errors.New("unauthorized to update this sale")
	}

	// Check if sale is already active
	now := time.Now().Unix()
	if s.isSaleActive(sale, now) {
		return nil, errors.New("cannot update active sale")
	}

	// Validate new dates if provided
	startDate := sale.StartDate
	endDate := sale.EndDate

	if req.StartDate != 0 {
		if req.StartDate <= now {
			return nil, errors.New("sale start date must be in the future")
		}
		startDate = req.StartDate
	}

	if req.EndDate != 0 {
		endDate = req.EndDate
	}

	if endDate <= startDate {
		return nil, errors.New("sale end date must be after start date")
	}

	// Check for overlapping sales (excluding current sale)
	existingSales, err := s.saleRepo.ListByEvent(sale.EventID)
	if err != nil {
		return nil, errors.New("failed to check existing sales")
	}

	for _, existingSale := range existingSales {
		if existingSale.ID != saleID && s.datesOverlap(startDate, endDate, existingSale.StartDate, existingSale.EndDate) {
			return nil, errors.New("sale dates overlap with existing sale")
		}
	}

	// Update fields
	if req.StartDate != 0 {
		sale.StartDate = req.StartDate
	}
	if req.EndDate != 0 {
		sale.EndDate = req.EndDate
	}

	if err := s.saleRepo.Update(sale); err != nil {
		return nil, errors.New("failed to update sale")
	}

	return s.saleToResponse(sale, event), nil
}

func (s *SaleService) DeleteSale(saleID, sellerID uint) error {
	sale, err := s.saleRepo.GetByID(saleID)
	if err != nil {
		return errors.New("sale not found")
	}

	// Verify seller owns the event
	event, err := s.eventRepo.GetByID(sale.EventID)
	if err != nil {
		return errors.New("event not found")
	}
	if event.SellerID != sellerID {
		return errors.New("unauthorized to delete this sale")
	}

	// Check if sale is already active
	now := time.Now().Unix()
	if s.isSaleActive(sale, now) {
		return errors.New("cannot delete active sale")
	}

	// TODO: Check if any tickets are sold for this sale
	// This would require checking the Ticket model for sold tickets with this SaleID

	if err := s.saleRepo.Delete(saleID); err != nil {
		return errors.New("failed to delete sale")
	}

	return nil
}

// Helper functions

func (s *SaleService) saleToResponse(sale *models.Sale, event *models.Event) *SaleResponse {
	now := time.Now().Unix()

	response := &SaleResponse{
		ID:        sale.ID,
		StartDate: sale.StartDate,
		EndDate:   sale.EndDate,
		EventID:   sale.EventID,
		IsActive:  s.isSaleActive(sale, now),
	}

	if event != nil {
		response.EventInfo.Title = event.Title
		response.EventInfo.Description = event.Description
		response.EventInfo.Date = event.Date
		response.EventInfo.Address = event.Address
	}

	return response
}

func (s *SaleService) isSaleActive(sale *models.Sale, currentTime int64) bool {
	return currentTime >= sale.StartDate && currentTime <= sale.EndDate
}

func (s *SaleService) datesOverlap(start1, end1, start2, end2 int64) bool {
	return start1 < end2 && start2 < end1
}
