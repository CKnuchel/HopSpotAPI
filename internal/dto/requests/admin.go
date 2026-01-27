package requests

import (
	"hopSpotAPI/internal/domain"
)

type ListUsersRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	Limit    int    `form:"limit,default=50" binding:"min=1,max=100"`
	IsActive *bool  `form:"is_active"`
	Role     string `form:"role" binding:"omitempty,oneof=user admin"`
	Search   string `form:"search"`
}

type AdminUpdateUserRequest struct {
	Role     *domain.Role `json:"role" binding:"omitempty,oneof=user admin"`
	IsActive *bool        `json:"is_active"`
}

type ListInvitationCodesRequest struct {
	Page       int   `form:"page,default=1" binding:"min=1"`
	Limit      int   `form:"limit,default=50" binding:"min=1,max=100"`
	IsRedeemed *bool `form:"is_redeemed"`
}

type CreateInvitationCodeRequest struct {
	Comment string `json:"comment" binding:"max=255"`
}
