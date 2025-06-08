// internal/handlers/admin_handler.go
package handlers

import (
	"strconv"

	"eticketing/internal/middleware"
	"eticketing/internal/models"
	"eticketing/internal/services"
	"eticketing/internal/utils"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService *services.AdminService
}

func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) GetProfile(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	profile, err := h.adminService.GetProfile(currentUser.UserID)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Admin profile retrieved successfully", profile)
}

func (h *AdminHandler) UpdateProfile(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	profile, err := h.adminService.UpdateProfile(currentUser.UserID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Admin profile updated successfully", profile)
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	var req services.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	err = h.adminService.ChangePassword(currentUser.UserID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Admin password changed successfully", nil)
}

func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	if currentUser.UserType != models.UserTypeAdmin {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	stats, err := h.adminService.GetSystemStats()
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "System statistics retrieved successfully", stats)
}

func (h *AdminHandler) GetPendingEvents(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	if currentUser.UserType != models.UserTypeAdmin {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	events, err := h.adminService.GetPendingEvents(page, limit)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Pending events retrieved successfully", events)
}

func (h *AdminHandler) ApproveEvent(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	if currentUser.UserType != models.UserTypeAdmin {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	err = h.adminService.ApproveEvent(uint(eventID))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Event approved successfully", nil)
}

func (h *AdminHandler) RejectEvent(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	if currentUser.UserType != models.UserTypeAdmin {
		utils.ForbiddenResponse(c, "Admin access required")
		return
	}

	eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	err = h.adminService.RejectEvent(uint(eventID), req.Reason)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Event rejected successfully", nil)
}
