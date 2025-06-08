// internal/handlers/user_handler.go
package handlers

import (
	"eticketing/internal/middleware"
	"eticketing/internal/services"
	"eticketing/internal/utils"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	profile, err := h.userService.GetProfile(currentUser.UserID)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Profile retrieved successfully", profile)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
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

	profile, err := h.userService.UpdateProfile(currentUser.UserID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Profile updated successfully", profile)
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
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

	err = h.userService.ChangePassword(currentUser.UserID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Password changed successfully", nil)
}

func (h *UserHandler) DeleteAccount(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	err = h.userService.DeleteAccount(currentUser.UserID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Account deleted successfully", nil)
}
