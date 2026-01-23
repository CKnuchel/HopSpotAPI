package domain

import "gorm.io/gorm"

type User struct {
	*gorm.Model
	Email        string  `gorm:"type:varchar(255);unique_index" json:"email"`
	PasswordHast string  `gorm:"type:varchar(255)" json:"-"`
	DisplayName  string  `gorm:"type:varchar(255)" json:"display_name"`
	Role         string  `gorm:"type:varchar(50)" json:"role"`
	FcmToken     *string `gorm:"type:varchar(255)" json:"fcm_token"`
	IsActive     bool    `gorm:"type:boolean" json:"is_active"`
}
