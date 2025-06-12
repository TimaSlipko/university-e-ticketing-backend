package handlers

import (
	"strconv"

	"eticketing/internal/middleware"
	"eticketing/internal/models"
	"eticketing/internal/services"
	"eticketing/internal/utils"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
}

func NewPaymentHandler(paymentService *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	var req services.PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	req.UserID = currentUser.UserID
	response, err := h.paymentService.ProcessPayment(&req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, "Payment processed successfully", response)
}

func (h *PaymentHandler) GetUserPayments(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	payments, err := h.paymentService.GetUserPayments(currentUser.UserID, models.UserTypeUser, limit, offset)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Payments retrieved successfully", payments)
}

func (h *PaymentHandler) GetSellerPayments(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	// Only sellers can access this endpoint
	if currentUser.UserType != models.UserTypeSeller {
		utils.ForbiddenResponse(c, "Only sellers can access seller payments")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Get seller revenue payments
	payments, err := h.paymentService.GetUserPayments(currentUser.UserID, models.UserTypeSeller, limit, offset)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Seller payments retrieved successfully", payments)
}

func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	paymentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid payment ID")
		return
	}

	response, err := h.paymentService.GetPaymentStatus(uint(paymentID))
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Payment status retrieved successfully", response)
}

func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	// Only admins can process refunds
	if currentUser.UserType != models.UserTypeAdmin {
		utils.ForbiddenResponse(c, "Admin access required for refunds")
		return
	}

	paymentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid payment ID")
		return
	}

	err = h.paymentService.RefundPayment(uint(paymentID))
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Payment refunded successfully", nil)
}
