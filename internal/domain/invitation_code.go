package domain

import (
	"gorm.io/gorm"
)

type InvitationCode struct {
	*gorm.Model
	Code       string `gorm:"unique;type:varchar(100)" json:"code"`
	Comment    string `gorm:"type:varchar(255)" json:"comment"`
	CreatedBy  uint   `gorm:"type:int;not null" json:"createdBy"`
	RedeemedBy uint   `gorm:"type:int;default:null" json:"redeemedBy,omitempty"`

	// Relations - loaded with Preload
	Creator  User  `gorm:"foreignKey:CreatedBy;references:ID" json:"creator,omitempty"`
	Redeemer *User `gorm:"foreignKey:RedeemedBy;references:ID" json:"redeemer,omitempty"`
}
