package domain

import (
	"time"
)

// ActionType constants for activity types
const (
	ActionBenchCreated  = "bench_created"
	ActionVisitAdded    = "visit_added"
	ActionFavoriteAdded = "favorite_added"
)

type Activity struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index:idx_activity_created,priority:2" json:"userId"`
	ActionType string    `gorm:"type:varchar(50);index:idx_activity_type" json:"actionType"`
	BenchID    *uint     `gorm:"index" json:"benchId,omitempty"`
	CreatedAt  time.Time `gorm:"type:timestamptz;index:idx_activity_created,priority:1" json:"createdAt"`

	// Relations - loaded with Preload
	User  User   `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Bench *Bench `gorm:"foreignKey:BenchID;references:ID" json:"bench,omitempty"`
}
