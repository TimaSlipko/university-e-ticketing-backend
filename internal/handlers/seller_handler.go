// internal/handlers/seller_handler.go
package handlers

import (
	"eticketing/internal/middleware"
	"eticketing/internal/services"
	"eticketing/internal/utils"
	"github.com/gin-gonic/gin"
)

type SellerHandler struct {
	sellerService *services.SellerService
}

func NewSellerHandler(sellerService *services.SellerService) *SellerHandler {
	return &SellerHandler{sellerService: sellerService}
}

func (h *SellerHandler) GetProfile(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	profile, err := h.sellerService.GetProfile(currentUser.UserID)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Seller profile retrieved successfully", profile)
}

func (h *SellerHandler) UpdateProfile(c *gin.Context) {
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

	profile, err := h.sellerService.UpdateProfile(currentUser.UserID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Seller profile updated successfully", profile)
}

func (h *SellerHandler) ChangePassword(c *gin.Context) {
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

	err = h.sellerService.ChangePassword(currentUser.UserID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Seller password changed successfully", nil)
}

func (h *SellerHandler) GetStats(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	stats, err := h.sellerService.GetSellerStats(currentUser.UserID)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Seller statistics retrieved successfully", stats)
}

func (h *SellerHandler) DeleteAccount(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	err = h.sellerService.DeleteAccount(currentUser.UserID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Seller account deleted successfully", nil)
}
