package service

import (
	"context"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/apperror"
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
	jwtSecret      string
}

// NewAuthService Constructor
func NewAuthService(userRepo repository.UserRepository, invitationRepo repository.InvitationRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo:       userRepo,
		invitationRepo: invitationRepo,
		jwtSecret:      jwtSecret,
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
	panic("implement me") //TODO: Implement password hashing and co
}

func (s authService) Login(ctx context.Context, req *requests.LoginRequest) (*responses.LoginResponse, error) {
	panic("implement me")
}
