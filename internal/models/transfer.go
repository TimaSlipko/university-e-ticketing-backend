package models

type TransferStatus int

const (
	TransferStatusPending   TransferStatus = 1
	TransferStatusAccepted  TransferStatus = 2
	TransferStatusRejected  TransferStatus = 3
	TransferStatusCancelled TransferStatus = 4
)

type ActiveTicketTransfer struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	FromUserID        uint           `json:"from_user_id" gorm:"not null"`
	ToUserID          uint           `json:"to_user_id" gorm:"not null"`
	Date              int64          `json:"date" gorm:"not null"` // Unix timestamp
	PurchasedTicketID uint           `json:"purchased_ticket_id" gorm:"not null"`
	Status            TransferStatus `json:"status" gorm:"default:1"`

	// Relationships
	FromUser        User            `json:"from_user" gorm:"foreignKey:FromUserID"`
	ToUser          User            `json:"to_user" gorm:"foreignKey:ToUserID"`
	PurchasedTicket PurchasedTicket `json:"purchased_ticket" gorm:"foreignKey:PurchasedTicketID"`
}

type DoneTicketTransfer struct {
	ID                uint  `json:"id" gorm:"primaryKey"`
	FromUserID        uint  `json:"from_user_id" gorm:"not null"`
	ToUserID          uint  `json:"to_user_id" gorm:"not null"`
	Date              int64 `json:"date" gorm:"not null"` // Unix timestamp
	PurchasedTicketID uint  `json:"purchased_ticket_id" gorm:"not null"`
	CompletedAt       int64 `json:"completed_at" gorm:"not null"` // Unix timestamp

	// Relationships
	FromUser        User            `json:"from_user" gorm:"foreignKey:FromUserID"`
	ToUser          User            `json:"to_user" gorm:"foreignKey:ToUserID"`
	PurchasedTicket PurchasedTicket `json:"purchased_ticket" gorm:"foreignKey:PurchasedTicketID"`
}
