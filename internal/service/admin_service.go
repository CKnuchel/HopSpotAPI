package service

import (
	"context"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/apperror"
	"hopSpotAPI/pkg/utils"
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

func (a *adminService) ListUsers(ctx context.Context, req *requests.ListUsersRequest) (*responses.PaginatedUsersResponse, error) {
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

func (a *adminService) UpdateUser(ctx context.Context, id uint, req *requests.AdminUpdateUserRequest) (*responses.UserResponse, error) {
	// Get existing user
	user, err := a.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrUserNotFound
	}

	// Update fields
	if req.Role != nil {
		user.Role = *req.Role
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Update user
	if err := a.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	response := mapper.UserToResponse(user)
	return &response, nil
}

func (a *adminService) DeleteUser(ctx context.Context, id uint, adminID uint) error {
	// Admin cannot delete themselves
	if id == adminID {
		return apperror.ErrCannotDeleteSelf
	}

	user, err := a.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.ErrUserNotFound
	}

	return a.userRepo.Delete(ctx, id)
}

func (a *adminService) ListInvitationCodes(ctx context.Context, req *requests.ListInvitationCodesRequest) (*responses.PaginatedInvitationCodesResponse, error) {
	filter := repository.InvitationFilter{
		Page:       req.Page,
		Limit:      req.Limit,
		IsRedeemed: req.IsRedeemed,
	}

	invitationCodes, total, err := a.invitationCodeRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Pagination calculation
	totalPages := int(total) / req.Limit
	if int(total)%req.Limit > 0 {
		totalPages++
	}

	return &responses.PaginatedInvitationCodesResponse{
		Codes: mapper.InvitationCodesToResponse(invitationCodes),
		Pagination: responses.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// CreateInvitationCode implements [AdminService].
func (a *adminService) CreateInvitationCode(ctx context.Context, req *requests.CreateInvitationCodeRequest, adminID uint) (*responses.InvitationCodeResponse, error) {
	// Generate unique code
	code, err := utils.GenerateInvitationCode(6)
	if err != nil {
		return nil, err
	}

	// Create invitation code entity
	invitationCode := &domain.InvitationCode{
		Code:      code,
		Comment:   req.Comment,
		CreatedBy: &adminID,
	}

	// Save to repository
	if err := a.invitationCodeRepo.Create(ctx, invitationCode); err != nil {
		return nil, err
	}

	response := mapper.InvitationCodeToResponse(invitationCode)
	return &response, nil
}
