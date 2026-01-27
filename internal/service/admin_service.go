package service

import (
	"context"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
)

type AdminService interface {
	// Users
	ListUsers(ctx context.Context, req *requests.ListUsersRequest) (*responses.PaginatedUsersResponse, error)
	UpdateUser(ctx context.Context, id uint, req *requests.AdminUpdateUserRequest) (*responses.UserResponse, error)
	DeleteUser(ctx context.Context, id uint, adminID uint) error

	// Invitation Codes
	ListInvitationCodes(ctx context.Context, req *requests.ListInvitationCodesRequest) (*responses.PaginatedInvitationCodesResponse, error)
	CreateInvitationCode(ctx context.Context, req *requests.CreateInvitationCodeRequest, adminID uint) (*responses.InvitationCodeResponse, error)
}
