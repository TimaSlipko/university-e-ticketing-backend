// internal/models/ticket.go
package models

type TicketType int

const (
	TicketTypeRegular TicketType = 1
	TicketTypeVIP     TicketType = 2
	TicketTypePremium TicketType = 3
)

type Ticket struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Price       float64    `json:"price" gorm:"not null"`
	IsHeld      bool       `json:"is_held" gorm:"default:false"`
	IsSold      bool       `json:"is_sold" gorm:"default:false"`
	Type        TicketType `json:"type" gorm:"not null"`
	IsVip       bool       `json:"is_vip" gorm:"default:false"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description" gorm:"type:text"`
	Place       string     `json:"place" gorm:"not null"` // Seat/section info
	SaleID      uint       `json:"sale_id" gorm:"not null"`
	EventID     uint       `json:"event_id" gorm:"not null"` // Added for easier querying

	// Relationships
	Sale  Sale  `json:"sale" gorm:"foreignKey:SaleID"`
	Event Event `json:"event" gorm:"foreignKey:EventID"`
}

type PurchasedTicket struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Price       float64    `json:"price" gorm:"not null"`
	Type        TicketType `json:"type" gorm:"not null"`
	IsVip       bool       `json:"is_vip" gorm:"default:false"`
	Title       string     `json:"title" gorm:"not null"`
	Description string     `json:"description" gorm:"type:text"`
	Place       string     `json:"place" gorm:"not null"`
	UserID      uint       `json:"user_id" gorm:"not null"`
	TicketID    uint       `json:"ticket_id" gorm:"not null"`
	IsUsed      bool       `json:"is_used" gorm:"default:false"`
	UsedAt      *int64     `json:"used_at"` // Unix timestamp, nullable

	// Relationships
	User   User   `json:"user" gorm:"foreignKey:UserID"`
	Ticket Ticket `json:"ticket" gorm:"foreignKey:TicketID"`
}

// GroupedTicket represents aggregated ticket data for display purposes
type GroupedTicket struct {
	Price           float64    `json:"price"`
	Type            TicketType `json:"type"`
	IsVip           bool       `json:"is_vip"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	Place           string     `json:"place"`
	SaleID          uint       `json:"sale_id"`
	EventID         uint       `json:"event_id"`
	TotalAmount     int        `json:"total_amount"`
	AvailableAmount int        `json:"available_amount"`
	SoldAmount      int        `json:"sold_amount"`
	HeldAmount      int        `json:"held_amount"`
}
