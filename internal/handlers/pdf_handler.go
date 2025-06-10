// internal/handlers/pdf_handler.go
package handlers

import (
	"fmt"
	"strconv"

	"eticketing/internal/middleware"
	"eticketing/internal/repositories"
	"eticketing/internal/services"
	"eticketing/internal/utils"
	"github.com/gin-gonic/gin"
)

type PDFHandler struct {
	pdfService          *services.PDFService
	purchasedTicketRepo repositories.PurchasedTicketRepository
	eventRepo           repositories.EventRepository
}

func NewPDFHandler(
	pdfService *services.PDFService,
	purchasedTicketRepo repositories.PurchasedTicketRepository,
	eventRepo repositories.EventRepository,
) *PDFHandler {
	return &PDFHandler{
		pdfService:          pdfService,
		purchasedTicketRepo: purchasedTicketRepo,
		eventRepo:           eventRepo,
	}
}

func (h *PDFHandler) DownloadTicketPDF(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	ticketID, err := strconv.ParseUint(c.Param("ticket_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid ticket ID")
		return
	}

	// Get purchased ticket
	purchasedTicket, err := h.purchasedTicketRepo.GetByID(uint(ticketID))
	if err != nil {
		utils.NotFoundResponse(c, "Ticket not found")
		return
	}

	// Verify ticket ownership
	if purchasedTicket.UserID != currentUser.UserID {
		utils.ForbiddenResponse(c, "You can only download your own tickets")
		return
	}

	// Get event information
	event, err := h.eventRepo.GetByID(purchasedTicket.Ticket.EventID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to load event information")
		return
	}

	// Prepare PDF data
	pdfData := &services.TicketPDFData{
		PurchasedTicket: purchasedTicket,
		Event:           event,
		QRCodeURL:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
	}

	// Generate PDF
	pdfBytes, err := h.pdfService.GenerateTicketPDF(pdfData)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate PDF: "+err.Error())
		return
	}

	// Set response headers for PDF download
	filename := fmt.Sprintf("ticket_%d_%s.pdf", purchasedTicket.ID, event.Title)
	// Sanitize filename for safe download
	filename = utils.SanitizeFilename(filename)

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	// Write PDF to response
	c.Data(200, "application/pdf", pdfBytes)
}

func (h *PDFHandler) ViewTicketPDF(c *gin.Context) {
	currentUser, err := middleware.GetCurrentUser(c)
	if err != nil {
		utils.UnauthorizedResponse(c, "Unauthorized")
		return
	}

	ticketID, err := strconv.ParseUint(c.Param("ticket_id"), 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid ticket ID")
		return
	}

	// Get purchased ticket
	purchasedTicket, err := h.purchasedTicketRepo.GetByID(uint(ticketID))
	if err != nil {
		utils.NotFoundResponse(c, "Ticket not found")
		return
	}

	// Verify ticket ownership
	if purchasedTicket.UserID != currentUser.UserID {
		utils.ForbiddenResponse(c, "You can only view your own tickets")
		return
	}

	// Get event information
	event, err := h.eventRepo.GetByID(purchasedTicket.Ticket.EventID)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to load event information")
		return
	}

	// Prepare PDF data
	pdfData := &services.TicketPDFData{
		PurchasedTicket: purchasedTicket,
		Event:           event,
		QRCodeURL:       "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
	}

	// Generate PDF
	pdfBytes, err := h.pdfService.GenerateTicketPDF(pdfData)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to generate PDF: "+err.Error())
		return
	}

	// Set response headers for PDF viewing in browser
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "inline")
	c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))

	// Write PDF to response
	c.Data(200, "application/pdf", pdfBytes)
}
