package models

type User struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Username     string `json:"username" gorm:"unique;not null"`
	PasswordHash string `json:"-" gorm:"not null"`
	Email        string `json:"email" gorm:"unique;not null"`
	Name         string `json:"name" gorm:"not null"`
	Surname      string `json:"surname" gorm:"not null"`

	// Relationships
	PurchasedTickets []PurchasedTicket `json:"purchased_tickets,omitempty" gorm:"foreignKey:UserID"`
	PaymentMethods   []PaymentMethod   `json:"payment_methods,omitempty" gorm:"foreignKey:UserID"`
	//Payments         []Payment         `json:"payments,omitempty" gorm:"foreignKey:UserID"`
}

type Seller struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Username     string `json:"username" gorm:"unique;not null"`
	PasswordHash string `json:"-" gorm:"not null"`
	Email        string `json:"email" gorm:"unique;not null"`
	Name         string `json:"name" gorm:"not null"`
	Surname      string `json:"surname" gorm:"not null"`

	// Relationships
	Events []Event `json:"events,omitempty" gorm:"foreignKey:SellerID"`
}

type Admin struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	Username     string `json:"username" gorm:"unique;not null"`
	PasswordHash string `json:"-" gorm:"not null"`
	Email        string `json:"email" gorm:"unique;not null"`
	Name         string `json:"name" gorm:"not null"`
	Surname      string `json:"surname" gorm:"not null"`
	AdminRole    int    `json:"admin_role" gorm:"default:1"` // 1=regular admin, 2=super admin
}

type UserType int

const (
	UserTypeUser   UserType = 1
	UserTypeSeller UserType = 2
	UserTypeAdmin  UserType = 3
)
