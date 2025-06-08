package models

type PaymentType int
type PaymentStatus int

const (
	PaymentTypeCard      PaymentType = 1
	PaymentTypePayPal    PaymentType = 2
	PaymentTypeGooglePay PaymentType = 3
	PaymentTypeStripe    PaymentType = 4
)

const (
	PaymentStatusPending   PaymentStatus = 1
	PaymentStatusCompleted PaymentStatus = 2
	PaymentStatusFailed    PaymentStatus = 3
	PaymentStatusRefunded  PaymentStatus = 4
)

type Payment struct {
	ID     uint          `json:"id" gorm:"primaryKey"`
	UserID uint          `json:"user_id" gorm:"not null"`
	Date   int64         `json:"date" gorm:"not null"` // Unix timestamp
	Type   PaymentType   `json:"type" gorm:"not null"`
	Amount float64       `json:"amount" gorm:"not null"`
	Status PaymentStatus `json:"status" gorm:"default:1"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

type PaymentMethod struct {
	ID        uint        `json:"id" gorm:"primaryKey"`
	Type      PaymentType `json:"type" gorm:"not null"`
	Token     string      `json:"token" gorm:"not null"`  // Encrypted token
	Data      string      `json:"data" gorm:"type:jsonb"` // Additional payment data as JSON
	UserID    uint        `json:"user_id" gorm:"not null"`
	IsDefault bool        `json:"is_default" gorm:"default:false"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}
