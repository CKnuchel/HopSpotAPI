package domain

import (
	"time"
)

type Favorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_spot,priority:1" json:"userId"`
	SpotID    uint      `gorm:"uniqueIndex:idx_user_spot,priority:2" json:"spotId"`
	CreatedAt time.Time `json:"createdAt"`

	// Relations - loaded with Preload
	Spot Spot `gorm:"foreignKey:SpotID;references:ID" json:"spot,omitempty"`
	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}
