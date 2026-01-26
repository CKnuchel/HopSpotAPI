package service

import (
	"context"
	"hopSpotAPI/internal/config"
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/apperror"
	"hopSpotAPI/pkg/utils"
)

// AuthService Interface defining authentication service methods
type AuthService interface {
	Register(ctx context.Context, req *requests.RegisterRequest) (*responses.LoginResponse, error)
	Login(ctx context.Context, req *requests.LoginRequest) (*responses.LoginResponse, error)
}

// Implementation
type authService struct {
	userRepo       repository.UserRepository
	invitationRepo repository.InvitationRepository
	config         config.Config
}

// NewAuthService Constructor
func NewAuthService(userRepo repository.UserRepository, invitationRepo repository.InvitationRepository, config config.Config) AuthService {
	return &authService{
		userRepo:       userRepo,
		invitationRepo: invitationRepo,
		config:         config,
	}
}

func (s authService) Register(ctx context.Context, req *requests.RegisterRequest) (*responses.LoginResponse, error) {

	// Check if email already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, apperror.ErrEmailAlreadyExists
	}

	// Check if invitation code is valid
	invitation, err := s.invitationRepo.FindByCode(ctx, req.InvitationCode)
	if err != nil {
		return nil, err
	}
	if invitation == nil {
		return nil, apperror.ErrInvalidInvitationCode
	}
	if invitation.RedeemedBy != nil {
		return nil, apperror.ErrInvitationCodeAlreadyRedeemed
	}

	// Hashing password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Mapping DTO -> User domain model
	user := mapper.RegisterRequestToUser(req)
	user.PasswordHash = hashedPassword
	user.Role = domain.RoleUser // Default role
	user.IsActive = true

	// Creating user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Mark invitation as redeemed
	if err := s.invitationRepo.MarkAsRedeemed(ctx, invitation.ID, user.ID); err != nil {
		return nil, err
	}

	// Generating JWT token
	token, err := utils.GenerateJWT(user, &s.config)
	if err != nil {
		return nil, err
	}

	// Preparing response
	return &responses.LoginResponse{
		User:  mapper.UserToResponse(user),
		Token: token,
	}, nil
}

func (s authService) Login(ctx context.Context, req *requests.LoginRequest) (*responses.LoginResponse, error) {

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.ErrInvalidCredentials
	}

	// Check if the user is active
	if !user.IsActive {
		return nil, apperror.ErrAccountDeactivated
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, apperror.ErrInvalidCredentials
	}

	// Generating JWT token
	token, err := utils.GenerateJWT(user, &s.config)
	if err != nil {
		return nil, err
	}

	// Preparing response
	return &responses.LoginResponse{
		User:  mapper.UserToResponse(user),
		Token: token,
	}, nil
}
