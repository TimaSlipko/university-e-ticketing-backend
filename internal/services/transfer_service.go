// internal/services/transfer_service.go
package services

import (
	"errors"
	"time"

	"eticketing/internal/models"
	"eticketing/internal/repositories"
	"gorm.io/gorm"
)

type TransferService struct {
	transferRepo        repositories.TransferRepository
	purchasedTicketRepo repositories.PurchasedTicketRepository
	userRepo            repositories.UserRepository
}

type InitiateTransferRequest struct {
	FromUserID        uint   `json:"-"` // Set by handler
	ToUserEmail       string `json:"to_user_email" binding:"required,email"`
	PurchasedTicketID uint   `json:"purchased_ticket_id" binding:"required"`
}

type TransferResponse struct {
	ID         uint                  `json:"id"`
	FromUser   UserInfo              `json:"from_user"`
	ToUser     UserInfo              `json:"to_user"`
	TicketInfo PurchasedTicketInfo   `json:"ticket_info"`
	Status     models.TransferStatus `json:"status"`
	Date       int64                 `json:"date"`
}

func NewTransferService(
	transferRepo repositories.TransferRepository,
	purchasedTicketRepo repositories.PurchasedTicketRepository,
	userRepo repositories.UserRepository,
) *TransferService {
	return &TransferService{
		transferRepo:        transferRepo,
		purchasedTicketRepo: purchasedTicketRepo,
		userRepo:            userRepo,
	}
}

func (s *TransferService) InitiateTransfer(req *InitiateTransferRequest) (*TransferResponse, error) {
	// Get purchased ticket
	purchasedTicket, err := s.purchasedTicketRepo.GetByID(req.PurchasedTicketID)
	if err != nil {
		return nil, errors.New("purchased ticket not found")
	}

	// Check if user owns the ticket
	if purchasedTicket.UserID != req.FromUserID {
		return nil, errors.New("unauthorized to transfer this ticket")
	}

	if purchasedTicket.IsUsed {
		return nil, errors.New("cannot transfer used ticket")
	}

	// Find recipient user by email
	toUser, err := s.userRepo.GetByEmail(req.ToUserEmail)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("recipient user not found")
		}
		return nil, errors.New("failed to find recipient user")
	}

	// Check if trying to transfer to self
	if toUser.ID == req.FromUserID {
		return nil, errors.New("cannot transfer ticket to yourself")
	}

	// Create active transfer
	transfer := &models.ActiveTicketTransfer{
		FromUserID:        req.FromUserID,
		ToUserID:          toUser.ID,
		Date:              time.Now().Unix(),
		PurchasedTicketID: req.PurchasedTicketID,
		Status:            models.TransferStatusPending,
	}

	if err := s.transferRepo.CreateActive(transfer); err != nil {
		return nil, errors.New("failed to create transfer request")
	}

	// Get from user info
	fromUser, _ := s.userRepo.GetByID(req.FromUserID)

	return &TransferResponse{
		ID: transfer.ID,
		FromUser: UserInfo{
			ID:       fromUser.ID,
			Username: fromUser.Username,
			Email:    fromUser.Email,
			Name:     fromUser.Name,
			Surname:  fromUser.Surname,
			UserType: models.UserTypeUser,
		},
		ToUser: UserInfo{
			ID:       toUser.ID,
			Username: toUser.Username,
			Email:    toUser.Email,
			Name:     toUser.Name,
			Surname:  toUser.Surname,
			UserType: models.UserTypeUser,
		},
		TicketInfo: PurchasedTicketInfo{
			ID:          purchasedTicket.ID,
			TicketID:    purchasedTicket.TicketID,
			Title:       purchasedTicket.Title,
			Description: purchasedTicket.Description,
			Place:       purchasedTicket.Place,
			Price:       purchasedTicket.Price,
			IsUsed:      purchasedTicket.IsUsed,
		},
		Status: transfer.Status,
		Date:   transfer.Date,
	}, nil
}

func (s *TransferService) GetActiveTransfers(userID uint) ([]TransferResponse, error) {
	transfers, err := s.transferRepo.ListActiveByUser(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve active transfers")
	}

	var responses []TransferResponse
	for _, transfer := range transfers {
		response := TransferResponse{
			ID: transfer.ID,
			FromUser: UserInfo{
				ID:       transfer.FromUser.ID,
				Username: transfer.FromUser.Username,
				Email:    transfer.FromUser.Email,
				Name:     transfer.FromUser.Name,
				Surname:  transfer.FromUser.Surname,
				UserType: models.UserTypeUser,
			},
			ToUser: UserInfo{
				ID:       transfer.ToUser.ID,
				Username: transfer.ToUser.Username,
				Email:    transfer.ToUser.Email,
				Name:     transfer.ToUser.Name,
				Surname:  transfer.ToUser.Surname,
				UserType: models.UserTypeUser,
			},
			TicketInfo: PurchasedTicketInfo{
				ID:          transfer.PurchasedTicket.ID,
				TicketID:    transfer.PurchasedTicket.TicketID,
				Title:       transfer.PurchasedTicket.Title,
				Description: transfer.PurchasedTicket.Description,
				Place:       transfer.PurchasedTicket.Place,
				Price:       transfer.PurchasedTicket.Price,
				IsUsed:      transfer.PurchasedTicket.IsUsed,
			},
			Status: transfer.Status,
			Date:   transfer.Date,
		}
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *TransferService) AcceptTransfer(transferID, userID uint) error {
	transfer, err := s.transferRepo.GetActiveByID(transferID)
	if err != nil {
		return errors.New("transfer not found")
	}

	// Check if user is the recipient
	if transfer.ToUserID != userID {
		return errors.New("unauthorized to accept this transfer")
	}

	if transfer.Status != models.TransferStatusPending {
		return errors.New("transfer is not in pending status")
	}

	// Update transfer status
	transfer.Status = models.TransferStatusAccepted
	if err := s.transferRepo.UpdateActive(transfer); err != nil {
		return errors.New("failed to update transfer status")
	}

	// Update ticket ownership
	purchasedTicket, err := s.purchasedTicketRepo.GetByID(transfer.PurchasedTicketID)
	if err != nil {
		return errors.New("failed to find purchased ticket")
	}

	purchasedTicket.UserID = transfer.ToUserID
	if err := s.purchasedTicketRepo.Update(purchasedTicket); err != nil {
		return errors.New("failed to transfer ticket ownership")
	}

	// Create done transfer record
	doneTransfer := &models.DoneTicketTransfer{
		FromUserID:        transfer.FromUserID,
		ToUserID:          transfer.ToUserID,
		Date:              transfer.Date,
		PurchasedTicketID: transfer.PurchasedTicketID,
		CompletedAt:       time.Now().Unix(),
	}

	if err := s.transferRepo.CreateDone(doneTransfer); err != nil {
		// Log error but don't fail the transfer
		// The main transfer is already complete
	}

	return nil
}

func (s *TransferService) RejectTransfer(transferID, userID uint) error {
	transfer, err := s.transferRepo.GetActiveByID(transferID)
	if err != nil {
		return errors.New("transfer not found")
	}

	// Check if user is the recipient
	if transfer.ToUserID != userID {
		return errors.New("unauthorized to reject this transfer")
	}

	if transfer.Status != models.TransferStatusPending {
		return errors.New("transfer is not in pending status")
	}

	// Update transfer status
	transfer.Status = models.TransferStatusRejected
	if err := s.transferRepo.UpdateActive(transfer); err != nil {
		return errors.New("failed to update transfer status")
	}

	return nil
}

func (s *TransferService) GetTransferHistory(userID uint) ([]TransferHistoryResponse, error) {
	// Get completed transfers from DoneTicketTransfer table
	doneTransfers, err := s.transferRepo.ListDoneByUser(userID)
	if err != nil {
		return nil, errors.New("failed to retrieve transfer history")
	}

	var responses []TransferHistoryResponse
	for _, transfer := range doneTransfers {
		response := TransferHistoryResponse{
			ID: transfer.ID,
			FromUser: UserInfo{
				ID:       transfer.FromUser.ID,
				Username: transfer.FromUser.Username,
				Email:    transfer.FromUser.Email,
				Name:     transfer.FromUser.Name,
				Surname:  transfer.FromUser.Surname,
				UserType: models.UserTypeUser,
			},
			ToUser: UserInfo{
				ID:       transfer.ToUser.ID,
				Username: transfer.ToUser.Username,
				Email:    transfer.ToUser.Email,
				Name:     transfer.ToUser.Name,
				Surname:  transfer.ToUser.Surname,
				UserType: models.UserTypeUser,
			},
			TicketInfo: PurchasedTicketInfo{
				ID:          transfer.PurchasedTicket.ID,
				TicketID:    transfer.PurchasedTicket.TicketID,
				Title:       transfer.PurchasedTicket.Title,
				Description: transfer.PurchasedTicket.Description,
				Place:       transfer.PurchasedTicket.Place,
				Price:       transfer.PurchasedTicket.Price,
				IsUsed:      transfer.PurchasedTicket.IsUsed,
			},
			Date:        transfer.Date,
			CompletedAt: transfer.CompletedAt,
		}
		responses = append(responses, response)
	}

	return responses, nil
}

type TransferHistoryResponse struct {
	ID          uint                `json:"id"`
	FromUser    UserInfo            `json:"from_user"`
	ToUser      UserInfo            `json:"to_user"`
	TicketInfo  PurchasedTicketInfo `json:"ticket_info"`
	Date        int64               `json:"date"`
	CompletedAt int64               `json:"completed_at"`
}
