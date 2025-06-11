package handlers

import (
	"strconv"

	"eticketing/internal/middleware"
	"eticketing/internal/services"
	"eticketing/internal/utils"
	"github.com/gin-gonic/gin"
)

type PaymentMethodHandler struct {
	paymentMethodService *services.PaymentMethodService
}

func NewPaymentMethodHandler(paymentMethodService *services.PaymentMethodService) *PaymentMethodHandler {
	return &PaymentMethodHandler{paymentMethodService: paymentMethodService}
}

func (h *PaymentMethodHandler) CreatePaymentMethod(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	var req services.CreatePaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	req.UserID = currentUser.UserID
	response, err := h.paymentMethodService.CreatePaymentMethod(&req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, "Payment method created successfully", response)
}

func (h *PaymentMethodHandler) GetPaymentMethods(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	methods, err := h.paymentMethodService.GetUserPaymentMethods(currentUser.UserID)
	if err != nil {
		utils.InternalErrorResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Payment methods retrieved successfully", methods)
}

func (h *PaymentMethodHandler) GetPaymentMethod(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	methodID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid payment method ID")
		return
	}

	method, err := h.paymentMethodService.GetPaymentMethod(uint(methodID), currentUser.UserID)
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Payment method retrieved successfully", method)
}

func (h *PaymentMethodHandler) UpdatePaymentMethod(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	methodID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid payment method ID")
		return
	}

	var req services.UpdatePaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	err = h.paymentMethodService.UpdatePaymentMethod(uint(methodID), currentUser.UserID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Payment method updated successfully", nil)
}

func (h *PaymentMethodHandler) DeletePaymentMethod(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	methodID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid payment method ID")
		return
	}

	err = h.paymentMethodService.DeletePaymentMethod(uint(methodID), currentUser.UserID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Payment method deleted successfully", nil)
}

func (h *PaymentMethodHandler) SetDefaultPaymentMethod(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	methodID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid payment method ID")
		return
	}

	err = h.paymentMethodService.SetDefaultPaymentMethod(uint(methodID), currentUser.UserID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Default payment method set successfully", nil)
}
