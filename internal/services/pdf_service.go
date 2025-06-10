// internal/services/pdf_service.go
package services

import (
	"bytes"
	"fmt"
	"time"

	"eticketing/internal/models"
	"github.com/go-pdf/fpdf"
	"github.com/skip2/go-qrcode"
)

type PDFService struct{}

type TicketPDFData struct {
	PurchasedTicket *models.PurchasedTicket
	Event           *models.Event
	QRCodeURL       string
}

func NewPDFService() *PDFService {
	return &PDFService{}
}

func (s *PDFService) GenerateTicketPDF(data *TicketPDFData) ([]byte, error) {
	// Create new PDF document
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set margins
	pdf.SetMargins(20, 20, 20)

	// Title
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(41, 128, 185) // Blue color
	pdf.Cell(170, 15, "E-TICKET")
	pdf.Ln(20)

	// Event Title
	pdf.SetFont("Arial", "B", 18)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(170, 10, data.Event.Title)
	pdf.Ln(15)

	// Ticket Information Section
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(52, 73, 94)
	pdf.Cell(170, 8, "TICKET INFORMATION")
	pdf.Ln(8) // Increased from 2 to 8 for consistency

	// Line under section title
	pdf.SetDrawColor(52, 73, 94)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(12) // Increased from 8 to 12 for consistency

	// Ticket details
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)

	// Ticket ID
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 6, "Ticket ID:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(130, 6, fmt.Sprintf("#%d", data.PurchasedTicket.ID))
	pdf.Ln(8)

	// Ticket Type
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 6, "Type:")
	pdf.SetFont("Arial", "", 11)
	typeText := s.getTicketTypeText(data.PurchasedTicket.Type)
	if data.PurchasedTicket.IsVip {
		typeText += " (VIP)"
	}
	pdf.Cell(130, 6, typeText)
	pdf.Ln(8)

	// Ticket Title
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 6, "Category:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(130, 6, data.PurchasedTicket.Title)
	pdf.Ln(8)

	// Seat/Place
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 6, "Seat/Section:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(130, 6, data.PurchasedTicket.Place)
	pdf.Ln(8)

	// Price
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 6, "Price:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(130, 6, fmt.Sprintf("$%.2f", data.PurchasedTicket.Price))
	pdf.Ln(15)

	// Event Information Section
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(52, 73, 94)
	pdf.Cell(170, 8, "EVENT DETAILS")
	pdf.Ln(8) // Increased from 2 to 8 for more space

	pdf.SetDrawColor(52, 73, 94)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(12) // Increased from 8 to 12 for more space after line

	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)

	// Event Date
	eventDate := time.Unix(data.Event.Date, 0)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 6, "Date:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(130, 6, eventDate.Format("Monday, January 2, 2006"))
	pdf.Ln(8)

	// Event Time
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 6, "Time:")
	pdf.SetFont("Arial", "", 11)
	pdf.Cell(130, 6, eventDate.Format("3:04 PM"))
	pdf.Ln(8)

	// Event Location
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(40, 6, "Location:")
	pdf.SetFont("Arial", "", 11)
	// Handle long addresses by splitting lines
	addressLines := s.splitText(data.Event.Address, 130)
	for i, line := range addressLines {
		if i == 0 {
			pdf.Cell(130, 6, line)
		} else {
			pdf.Ln(6)
			pdf.Cell(40, 6, "")
			pdf.Cell(130, 6, line)
		}
	}
	pdf.Ln(8)

	// Event Description (if available)
	if data.Event.Description != "" {
		pdf.SetFont("Arial", "B", 11)
		pdf.Cell(40, 6, "Description:")
		pdf.Ln(6)
		pdf.SetFont("Arial", "", 10)

		// Handle multiline description
		descLines := s.splitText(data.Event.Description, 170)
		for _, line := range descLines[:min(3, len(descLines))] { // Limit to 3 lines
			pdf.Cell(170, 5, line)
			pdf.Ln(5)
		}
		pdf.Ln(5)
	}

	// QR Code Section
	pdf.Ln(10)
	pdf.SetFont("Arial", "B", 14)
	pdf.SetTextColor(52, 73, 94)
	pdf.Cell(170, 8, "QR CODE")
	pdf.Ln(8) // Increased from 2 to 8 for consistency

	pdf.SetDrawColor(52, 73, 94)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(12) // Increased from 10 to 12 for consistency

	// Generate QR code
	qrCode, err := qrcode.Encode(data.QRCodeURL, qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %v", err)
	}

	// Add QR code to PDF
	qrReader := bytes.NewReader(qrCode)
	pdf.RegisterImageReader("qr", "PNG", qrReader)

	// Center the QR code
	qrSize := 50.0
	pageWidth := 210.0 // A4 width in mm
	qrX := (pageWidth - qrSize) / 2

	pdf.Image("qr", qrX, pdf.GetY(), qrSize, qrSize, false, "PNG", 0, "")
	pdf.Ln(60) // Increased from 55 to 60 for more space after QR code

	// QR Code instruction
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(100, 100, 100)

	// Footer
	pdf.SetY(260) // Position near bottom
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(150, 150, 150)

	// Terms and conditions
	pdf.Cell(170, 4, "Entry is subject to terms and conditions. Please arrive 30 minutes before event start time.")
	pdf.Ln(12) // Increased from 8 to 12 for more space before generation info

	// Generation info
	pdf.SetTextColor(200, 200, 200)
	pdf.Cell(85, 4, fmt.Sprintf("Generated on: %s", time.Now().Format("Jan 2, 2006 at 3:04 PM")))
	pdf.Cell(85, 4, "E-Ticketing System")

	// Return PDF as bytes
	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %v", err)
	}

	return buf.Bytes(), nil
}

func (s *PDFService) getTicketTypeText(ticketType models.TicketType) string {
	switch ticketType {
	case models.TicketTypeRegular:
		return "Regular"
	case models.TicketTypeVIP:
		return "VIP"
	case models.TicketTypePremium:
		return "Premium"
	default:
		return "Unknown"
	}
}

func (s *PDFService) splitText(text string, maxWidth float64) []string {
	// Simple text splitting - in a real implementation, you might want more sophisticated word wrapping
	const avgCharWidth = 2.5 // Approximate character width in mm for Arial 11pt
	maxChars := int(maxWidth / avgCharWidth)

	if len(text) <= maxChars {
		return []string{text}
	}

	var lines []string
	for len(text) > maxChars {
		// Find the last space before maxChars
		splitIndex := maxChars
		for i := maxChars - 1; i > 0; i-- {
			if text[i] == ' ' {
				splitIndex = i
				break
			}
		}

		lines = append(lines, text[:splitIndex])
		text = text[splitIndex:]
		if len(text) > 0 && text[0] == ' ' {
			text = text[1:] // Remove leading space
		}
	}

	if len(text) > 0 {
		lines = append(lines, text)
	}

	return lines
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
