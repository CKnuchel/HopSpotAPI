package domain

import "gorm.io/gorm"

type User struct {
	*gorm.Model
	Email        string  `gorm:"type:varchar(255);unique_index" json:"email"`
	PasswordHash string  `gorm:"type:varchar(255)" json:"-"`
	DisplayName  string  `gorm:"type:varchar(255)" json:"display_name"`
	Role         Role    `gorm:"type:varchar(20);not null;default:'user'" json:"role"`
	FcmToken     *string `gorm:"type:varchar(255)" json:"fcm_token"`
	IsActive     bool    `gorm:"type:boolean" json:"is_active"`
}
