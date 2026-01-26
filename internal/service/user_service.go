package service

import (
	"context"
	"hopSpotAPI/internal/config"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/apperror"
	"hopSpotAPI/pkg/utils"
)

type UserService interface {
	GetProfile(ctx context.Context, userID uint) (*responses.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uint, req *requests.UpdateProfileRequest) (*responses.UserResponse, error)
	ChangePassword(ctx context.Context, userID uint, req *requests.ChangePasswordRequest) error
}

type userService struct {
	userRepo repository.UserRepository
	config   config.Config
}

func NewUserService(userRepo repository.UserRepository, cfg config.Config) UserService {
	return &userService{userRepo: userRepo, config: cfg}
}

func (u *userService) GetProfile(ctx context.Context, userID uint) (*responses.UserResponse, error) {
	// Find the user by ID
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrUserNotFound
	}

	// Map to response DTO
	response := mapper.UserToResponse(user)
	return &response, nil
}

// UpdateProfile implements UserService.
func (u *userService) UpdateProfile(ctx context.Context, userID uint, req *requests.UpdateProfileRequest) (*responses.UserResponse, error) {
	// Find the user by ID
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrUserNotFound
	}

	// Update fields if provided
	if req.DisplayName != nil {
		user.DisplayName = *req.DisplayName
	}

	// Save the updated user
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Map to response DTO
	response := mapper.UserToResponse(user)
	return &response, nil
}

func (u *userService) ChangePassword(ctx context.Context, userID uint, req *requests.ChangePasswordRequest) error {
	// Find the user by ID
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return apperror.ErrUserNotFound
	}

	// Verify the old password
	if !utils.CheckPasswordHash(req.OldPassword, user.PasswordHash) {
		return apperror.ErrInvalidCredentials
	}

	// Hash the new password
	newHashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update the user's password
	user.PasswordHash = newHashedPassword
	return u.userRepo.Update(ctx, user)
}
