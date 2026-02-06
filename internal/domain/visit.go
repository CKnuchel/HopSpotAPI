package domain

import (
	"time"

	"gorm.io/gorm"
)

type Visit struct {
	*gorm.Model
	SpotID    uint      `gorm:"index:idx_spot,priority:1" json:"spotId"`
	UserID    uint      `gorm:"index:idx_user,priority:1" json:"userId"`
	VisitedAt time.Time `gorm:"type:timestamptz;not null" json:"visitedAt"`
	Comment   string    `gorm:"type:varchar(255);not null" json:"comment"`

	// Relations - loaded with Preload
	Spot Spot `gorm:"foreignKey:SpotID;references:ID" json:"spot,omitempty"`
	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}
