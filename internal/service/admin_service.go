package service

import (
	"context"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
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
	// Defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}

	// Load users from repository
	filter := repository.UserFilter{
		Page:     req.Page,
		Limit:    req.Limit,
		IsActive: req.IsActive,
		Role:     &req.Role,
		Search:   req.Search,
	}

	if req.Role != "" {
		filter.Role = &req.Role
	}

	users, total, err := a.userRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Pagination calculation
	totalPages := int(total) / req.Limit
	if int(total)%req.Limit > 0 {
		totalPages++
	}

	return &responses.PaginatedUsersResponse{
		Users: mapper.UsersToResponse(users),
		Pagination: responses.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
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
