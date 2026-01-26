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
}
