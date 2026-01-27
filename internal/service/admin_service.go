package service

import (
	"context"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/repository"
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

type adminService struct {
	userRepo           repository.UserRepository
	invitationCodeRepo repository.InvitationRepository
}

func NewAdminService(userRepo repository.UserRepository, invitationCodeRepo repository.InvitationRepository) AdminService {
	return &adminService{
		userRepo:           userRepo,
		invitationCodeRepo: invitationCodeRepo,
	}
}

// ListUsers implements [AdminService].
func (a *adminService) ListUsers(ctx context.Context, req *requests.ListUsersRequest) (*responses.PaginatedUsersResponse, error) {
	panic("unimplemented")
}

// UpdateUser implements [AdminService].
func (a *adminService) UpdateUser(ctx context.Context, id uint, req *requests.AdminUpdateUserRequest) (*responses.UserResponse, error) {
	panic("unimplemented")
}

// DeleteUser implements [AdminService].
func (a *adminService) DeleteUser(ctx context.Context, id uint, adminID uint) error {
	panic("unimplemented")
}

// ListInvitationCodes implements [AdminService].
func (a *adminService) ListInvitationCodes(ctx context.Context, req *requests.ListInvitationCodesRequest) (*responses.PaginatedInvitationCodesResponse, error) {
	panic("unimplemented")
}

// CreateInvitationCode implements [AdminService].
func (a *adminService) CreateInvitationCode(ctx context.Context, req *requests.CreateInvitationCodeRequest, adminID uint) (*responses.InvitationCodeResponse, error) {
	panic("unimplemented")
}
