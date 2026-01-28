package database

import (
	"fmt"
	"hopSpotAPI/internal/domain"
	"log"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	log.Println("Migrating database...")

	err := db.AutoMigrate(
		&domain.User{},
		&domain.Bench{},
		&domain.Photo{},
		&domain.Notification{},
		&domain.InvitationCode{},
		&domain.Visit{},
		&domain.RefreshToken{},
	)

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}
