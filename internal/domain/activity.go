package domain

import (
	"time"
)

// ActionType constants for activity types
const (
	ActionSpotCreated   = "spot_created"
	ActionVisitAdded    = "visit_added"
	ActionFavoriteAdded = "favorite_added"
)

type Activity struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index:idx_activity_created,priority:2" json:"userId"`
	ActionType string    `gorm:"type:varchar(50);index:idx_activity_type" json:"actionType"`
	SpotID     *uint     `gorm:"index" json:"spotId,omitempty"`
	CreatedAt  time.Time `gorm:"type:timestamptz;index:idx_activity_created,priority:1" json:"createdAt"`

	// Relations - loaded with Preload
	User User  `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	Spot *Spot `gorm:"foreignKey:SpotID;references:ID" json:"spot,omitempty"`
}
