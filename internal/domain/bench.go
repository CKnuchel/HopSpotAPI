package domain

import (
	"time"
)

type Bench struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Name        string    `gorm:"type:varchar(255);uniqueIndex" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Latitude    float64   `gorm:"type:float;index:idx_location,priority:1" json:"latitude"`
	Longitude   float64   `gorm:"type:float;index:idx_location,priority:2" json:"longitude"`
	Rating      *int      `gorm:"type:int;default:null;index" json:"rating,omitempty"`
	HasToilet   bool      `gorm:"type:boolean" json:"has_toilet"`
	HasTrashBin bool      `gorm:"type:boolean" json:"has_trash_bin"`
	CreatedBy   uint      `gorm:"type:int;not null;index" json:"createdBy"`

	// Relations - loaded with Preload
	Creator User `gorm:"foreignKey:CreatedBy;references:ID" json:"creator"`
}
