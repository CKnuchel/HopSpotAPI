package domain

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	*gorm.Model
	UserID         uint       `gorm:"not null;index:idx_user_notifications,priority:1" json:"userId"`
	Type           string     `gorm:"type:varchar(255);not null" json:"type"`
	Title          string     `gorm:"type:varchar(255);not null" json:"title"`
	Message        string     `gorm:"type:text;not null" json:"message"`
	RelatedSpotID *uint      `gorm:"default:null" json:"relatedSpotId,omitempty"`
	RelatedUserID  *uint      `gorm:"default:null" json:"relatedUserId,omitempty"`
	IsRead         bool       `gorm:"default:false;index:idx_user_notifications,priority:2" json:"isRead"`
	SentAt         time.Time  `gorm:"not null;index:idx_user_notifications,priority:3" json:"sentAt"`
	ReadAt         *time.Time `gorm:"default:null" json:"readAt,omitempty"`

	// Relations - loaded with Preload
	User         User   `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	RelatedSpot *Spot `gorm:"foreignKey:RelatedSpotID;references:ID" json:"relatedSpot,omitempty"`
	RelatedUser  *User  `gorm:"foreignKey:RelatedUserID;references:ID" json:"relatedUser,omitempty"`
}
