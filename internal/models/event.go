package models

type EventStatus int

const (
	EventStatusPending   EventStatus = 1
	EventStatusApproved  EventStatus = 2
	EventStatusRejected  EventStatus = 3
	EventStatusCancelled EventStatus = 4
)

type Event struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	Title       string      `json:"title" gorm:"not null"`
	Description string      `json:"description" gorm:"type:text"`
	Date        int64       `json:"date" gorm:"not null"` // Unix timestamp
	Address     string      `json:"address" gorm:"not null"`
	Data        string      `json:"data" gorm:"type:json"` // Additional event data as JSON
	SellerID    uint        `json:"seller_id" gorm:"not null"`
	Status      EventStatus `json:"status" gorm:"default:1"`

	// Relationships
	Seller  Seller   `json:"seller" gorm:"foreignKey:SellerID"`
	Tickets []Ticket `json:"tickets,omitempty" gorm:"foreignKey:EventID"`
	Sales   []Sale   `json:"sales,omitempty" gorm:"foreignKey:EventID"`
}

type Sale struct {
	ID        uint  `json:"id" gorm:"primaryKey"`
	StartDate int64 `json:"start_date" gorm:"not null"` // Unix timestamp
	EndDate   int64 `json:"end_date" gorm:"not null"`   // Unix timestamp
	EventID   uint  `json:"event_id" gorm:"not null"`

	// Relationships
	Event Event `json:"event" gorm:"foreignKey:EventID"`
}
