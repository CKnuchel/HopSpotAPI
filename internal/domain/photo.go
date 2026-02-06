package domain

import (
	"gorm.io/gorm"
)

type Photo struct {
	*gorm.Model
	UploadedBy        uint   `gorm:"type:int;not null;index" json:"uploadedBy"`
	SpotID            uint   `gorm:"type:int;index:idx_spot_main,priority:1" json:"spotId"`
	IsMain            bool   `gorm:"type:boolean;default:false;index:idx_spot_main,priority:2" json:"isMain"`
	FilePathOriginal  string `gorm:"type:varchar(255);not null" json:"filePathOriginal"`
	FilePathMedium    string `gorm:"type:varchar(255);not null" json:"filePathMedium"`
	FilePathThumbnail string `gorm:"type:varchar(255);not null" json:"filePathThumbnail"`
	MimeType          string `gorm:"type:varchar(50);not null" json:"mimeType"`
	FileSize          int    `gorm:"type:int;not null" json:"fileSize"`

	// Relations - loaded with Preload
	Uploader User `gorm:"foreignKey:UploadedBy;references:ID" json:"uploader,omitempty"`
	Spot     Spot `gorm:"foreignKey:SpotID;references:ID" json:"spot,omitempty"`
}
