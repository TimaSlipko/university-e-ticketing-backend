package handlers

import (
	"strconv"

	"eticketing/internal/middleware"
	"eticketing/internal/models"
	"eticketing/internal/services"
	"eticketing/internal/utils"
	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	ticketService *services.TicketService
}

func NewTicketHandler(ticketService *services.TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ticketService}
}

func (h *TicketHandler) CreateTickets(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	if currentUser.UserType != models.UserTypeSeller {
		utils.ForbiddenResponse(c, "Only sellers can create tickets")
		return
	}

	var req services.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	err = h.ticketService.CreateTickets(&req, currentUser.UserID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, "Tickets created successfully", nil)
}

func (h *TicketHandler) UpdateTickets(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	if currentUser.UserType != models.UserTypeSeller {
		utils.ForbiddenResponse(c, "Only sellers can update tickets")
		return
	}

	eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	var reqBody struct {
		OldTicket models.GroupedTicket         `json:"old_ticket" binding:"required"`
		Updates   services.UpdateTicketRequest `json:"updates" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	err = h.ticketService.UpdateTickets(uint(eventID), currentUser.UserID, reqBody.OldTicket, &reqBody.Updates)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Tickets updated successfully", nil)
}

func (h *TicketHandler) DeleteTickets(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	if currentUser.UserType != models.UserTypeSeller {
		utils.ForbiddenResponse(c, "Only sellers can delete tickets")
		return
	}

	eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	var groupedTicket models.GroupedTicket
	if err := c.ShouldBindJSON(&groupedTicket); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	err = h.ticketService.DeleteTickets(uint(eventID), currentUser.UserID, groupedTicket)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Tickets deleted successfully", nil)
}

func (h *TicketHandler) GetGroupedEventTickets(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	tickets, err := h.ticketService.GetGroupedTicketsByEvent(uint(eventID))
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Grouped tickets retrieved successfully", tickets)
}

// Public endpoints

func (h *TicketHandler) GetAvailableGroupedEventTickets(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	tickets, err := h.ticketService.GetAvailableGroupedTicketsByEvent(uint(eventID))
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Available grouped tickets retrieved successfully", tickets)
}

// User endpoints - Ticket purchasing

func (h *TicketHandler) PurchaseTicketFromGroup(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	var req services.PurchaseTicketFromGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	req.UserID = currentUser.UserID
	response, err := h.ticketService.PurchaseTicketFromGroup(&req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, "Tickets purchased successfully", response)
}

func (h *TicketHandler) PurchaseTicket(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	var req services.PurchaseTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	req.UserID = currentUser.UserID
	response, err := h.ticketService.PurchaseTicket(&req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, "Ticket purchased successfully", response)
}

func (h *TicketHandler) GetMyTickets(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	tickets, err := h.ticketService.GetUserTickets(currentUser.UserID)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Tickets retrieved successfully", tickets)
}

func (h *TicketHandler) GetEventTickets(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("eventId"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	tickets, err := h.ticketService.GetEventTickets(uint(eventID))
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Event tickets retrieved successfully", tickets)
}

func (h *TicketHandler) TransferTicket(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	var req services.TransferTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	req.FromUserID = currentUser.UserID
	err = h.ticketService.TransferTicket(&req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Ticket transfer initiated successfully", nil)
}
