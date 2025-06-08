// internal/handlers/event_handler.go
package handlers

import (
	"strconv"

	"eticketing/internal/middleware"
	"eticketing/internal/models"
	"eticketing/internal/services"
	"eticketing/internal/utils"
	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventService *services.EventService
}

func NewEventHandler(eventService *services.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	if currentUser.UserType != models.UserTypeSeller {
		utils.ForbiddenResponse(c, "Only sellers can create events")
		return
	}

	var req services.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	req.SellerID = currentUser.UserID
	event, err := h.eventService.CreateEvent(&req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, "Event created successfully", event)
}

func (h *EventHandler) GetEvents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var events interface{}
	var err error

	events, err = h.eventService.GetEvents(page, limit)

	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Events retrieved successfully", events)
}

func (h *EventHandler) GetEvent(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	event, err := h.eventService.GetEventByID(uint(eventID))
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Event retrieved successfully", event)
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	var req services.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	event, err := h.eventService.UpdateEvent(uint(eventID), currentUser.UserID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Event updated successfully", event)
}

func (h *EventHandler) DeleteEvent(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	err = h.eventService.DeleteEvent(uint(eventID), currentUser.UserID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Event deleted successfully", nil)
}

func (h *EventHandler) GetMyEvents(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	if currentUser.UserType != models.UserTypeSeller {
		utils.ForbiddenResponse(c, "Only sellers can access this endpoint")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	events, err := h.eventService.GetEventsBySeller(currentUser.UserID, page, limit)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Events retrieved successfully", events)
}
