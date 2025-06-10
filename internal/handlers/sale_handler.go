package handlers

import (
	"strconv"

	"eticketing/internal/middleware"
	"eticketing/internal/models"
	"eticketing/internal/services"
	"eticketing/internal/utils"
	"github.com/gin-gonic/gin"
)

type SaleHandler struct {
	saleService *services.SaleService
}

func NewSaleHandler(saleService *services.SaleService) *SaleHandler {
	return &SaleHandler{saleService: saleService}
}

func (h *SaleHandler) CreateSale(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	if currentUser.UserType != models.UserTypeSeller {
		utils.ForbiddenResponse(c, "Only sellers can create sales")
		return
	}

	var req services.CreateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	sale, err := h.saleService.CreateSale(&req, currentUser.UserID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.CreatedResponse(c, "Sale created successfully", sale)
}

func (h *SaleHandler) GetSalesByEvent(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("event_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid event ID")
		return
	}

	sales, err := h.saleService.GetSalesByEvent(uint(eventID))
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Sales retrieved successfully", sales)
}

func (h *SaleHandler) GetSale(c *gin.Context) {
	saleID, err := strconv.ParseUint(c.Param("sale_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid sale ID")
		return
	}

	sale, err := h.saleService.GetSaleByID(uint(saleID))
	if err != nil {
		utils.NotFoundResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Sale retrieved successfully", sale)
}

func (h *SaleHandler) UpdateSale(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	if currentUser.UserType != models.UserTypeSeller {
		utils.ForbiddenResponse(c, "Only sellers can update sales")
		return
	}

	saleID, err := strconv.ParseUint(c.Param("sale_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid sale ID")
		return
	}

	var req services.UpdateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data")
		return
	}

	sale, err := h.saleService.UpdateSale(uint(saleID), currentUser.UserID, &req)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Sale updated successfully", sale)
}

func (h *SaleHandler) DeleteSale(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	if currentUser.UserType != models.UserTypeSeller {
		utils.ForbiddenResponse(c, "Only sellers can delete sales")
		return
	}

	saleID, err := strconv.ParseUint(c.Param("sale_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid sale ID")
		return
	}

	err = h.saleService.DeleteSale(uint(saleID), currentUser.UserID)
	if err != nil {
		utils.BadRequestResponse(c, err.Error())
		return
	}

	utils.SuccessResponse(c, "Sale deleted successfully", nil)
}
