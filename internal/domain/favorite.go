package domain

import (
	"time"
)

type Favorite struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_bench,priority:1" json:"userId"`
	BenchID   uint      `gorm:"uniqueIndex:idx_user_bench,priority:2" json:"benchId"`
	CreatedAt time.Time `json:"createdAt"`

	// Relations - loaded with Preload
	Bench Bench `gorm:"foreignKey:BenchID;references:ID" json:"bench,omitempty"`
	User  User  `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}
