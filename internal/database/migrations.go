package database

import (
	"log"

	"eticketing/internal/models"
)

func (d *Database) AutoMigrate() error {
	log.Println("Running database migrations...")

	err := d.DB.AutoMigrate(
		&models.Admin{},
		&models.User{},
		&models.Event{},
		&models.Sale{},
		&models.Ticket{},
		&models.PurchasedTicket{},
		&models.Payment{},
		&models.PaymentMethod{},
		&models.ActiveTicketTransfer{},
		&models.DoneTicketTransfer{},
	)

	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}
