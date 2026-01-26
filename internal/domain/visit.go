package domain

import (
	"time"

	"gorm.io/gorm"
)

type Visit struct {
	*gorm.Model
	BenchID   uint      `gorm:"index:idx_bench,priority:1" json:"benchId"`
	UserID    uint      `gorm:"index:idx_user,priority:1" json:"userId"`
	VisitedAt time.Time `gorm:"type:timestamptz;not null" json:"visitedAt"`
	Comment   string    `gorm:"type:varchar(255);not null" json:"comment"`

	// Relations - loaded with Preload
	Bench Bench `gorm:"foreignKey:BenchID;references:ID" json:"bench,omitempty"`
	User  User  `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}
