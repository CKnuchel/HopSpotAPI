package database

import (
	"fmt"
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/pkg/logger"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	logger.Info().Msg("Migrating database...")

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

	logger.Info().Msg("Migrations completed successfully")
	return nil
}
