// internal/handlers/transfer_handler.go
package handlers

import (
	"strconv"

	"eticketing/internal/middleware"
	"eticketing/internal/services"
	"eticketing/internal/utils"
	"github.com/gin-gonic/gin"
)

type TransferHandler struct {
	transferService *services.TransferService
}

func NewTransferHandler(transferService *services.TransferService) *TransferHandler {
	return &TransferHandler{transferService: transferService}
}

func (h *TransferHandler) InitiateTransfer(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	var req services.InitiateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	req.FromUserID = currentUser.UserID
	response, err := h.transferService.InitiateTransfer(&req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, "Transfer initiated successfully", response)
}

func (h *TransferHandler) GetActiveTransfers(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	transfers, err := h.transferService.GetActiveTransfers(currentUser.UserID)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Active transfers retrieved successfully", transfers)
}

func (h *TransferHandler) AcceptTransfer(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	transferID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid transfer ID")
		return
	}

	err = h.transferService.AcceptTransfer(uint(transferID), currentUser.UserID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Transfer accepted successfully", nil)
}

func (h *TransferHandler) RejectTransfer(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	transferID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid transfer ID")
		return
	}

	err = h.transferService.RejectTransfer(uint(transferID), currentUser.UserID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Transfer rejected successfully", nil)
}

func (h *TransferHandler) GetTransferHistory(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	history, err := h.transferService.GetTransferHistory(currentUser.UserID)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Transfer history retrieved successfully", history)
}
