package domain

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	*gorm.Model
	UserID    uint      `gorm:"not null;index" json:"userId"`
	TokenHash string    `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expiresAt"`
	IsRevoked bool      `gorm:"default:false" json:"isRevoked"`

	// Relation
	User User `gorm:"foreignKey:UserID;references:ID" json:"user"`
}

func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

func (r *RefreshToken) IsValid() bool {
	return !r.IsExpired() && !r.IsRevoked
}
