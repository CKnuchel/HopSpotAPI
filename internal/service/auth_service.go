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
	"time"
)

type AuthService interface {
	Register(ctx context.Context, req *requests.RegisterRequest) (*responses.LoginResponse, error)
	Login(ctx context.Context, req *requests.LoginRequest) (*responses.LoginResponse, error)
	Refresh(ctx context.Context, req *requests.RefreshTokenRequest) (*responses.LoginResponse, error)
	Logout(ctx context.Context, req *requests.LogoutRequest) error
	RefreshFCMToken(ctx context.Context, userId uint, fcmToken string) error
}

type authService struct {
	userRepo         repository.UserRepository
	invitationRepo   repository.InvitationRepository
	refreshTokenRepo repository.RefreshTokenRepository
	config           config.Config
}

func NewAuthService(
	userRepo repository.UserRepository,
	invitationRepo repository.InvitationRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	config config.Config,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		invitationRepo:   invitationRepo,
		refreshTokenRepo: refreshTokenRepo,
		config:           config,
	}
}

func (s *authService) Register(ctx context.Context, req *requests.RegisterRequest) (*responses.LoginResponse, error) {
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
	user.Role = domain.RoleUser
	user.IsActive = true

	// Creating user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Mark invitation as redeemed
	if err := s.invitationRepo.MarkAsRedeemed(ctx, invitation.ID, user.ID); err != nil {
		return nil, err
	}

	// Generate tokens
	return s.generateTokens(ctx, user)
}

func (s *authService) Login(ctx context.Context, req *requests.LoginRequest) (*responses.LoginResponse, error) {
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

	// Revoke all existing refresh tokens (single device policy)
	if err := s.refreshTokenRepo.RevokeByUserID(ctx, user.ID); err != nil {
		return nil, err
	}

	// Generate tokens
	return s.generateTokens(ctx, user)
}

func (s *authService) Refresh(ctx context.Context, req *requests.RefreshTokenRequest) (*responses.LoginResponse, error) {
	// Hash the provided token
	tokenHash := utils.HashToken(req.RefreshToken)

	// Find token in DB
	refreshToken, err := s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if refreshToken == nil {
		return nil, apperror.ErrInvalidRefreshToken
	}

	// Validate token
	if !refreshToken.IsValid() {
		return nil, apperror.ErrInvalidRefreshToken
	}

	// Check if user is still active
	if !refreshToken.User.IsActive {
		return nil, apperror.ErrAccountDeactivated
	}

	// Revoke current token (rotation)
	if err := s.refreshTokenRepo.RevokeByID(ctx, refreshToken.ID); err != nil {
		return nil, err
	}

	// Generate new tokens
	return s.generateTokens(ctx, &refreshToken.User)
}

func (s *authService) Logout(ctx context.Context, req *requests.LogoutRequest) error {
	// Hash the provided token
	tokenHash := utils.HashToken(req.RefreshToken)

	// Find token in DB
	refreshToken, err := s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return err
	}
	if refreshToken == nil {
		return nil // Token doesn't exist, already logged out
	}

	// Revoke the token
	return s.refreshTokenRepo.RevokeByID(ctx, refreshToken.ID)
}

func (s *authService) RefreshFCMToken(ctx context.Context, userId uint, fcmToken string) error {
	return s.userRepo.UpdateFCMToken(ctx, userId, fcmToken)
}

// generateTokens creates both access and refresh tokens
func (s *authService) generateTokens(ctx context.Context, user *domain.User) (*responses.LoginResponse, error) {
	// Generate Access Token (JWT)
	accessToken, err := utils.GenerateJWT(user, &s.config)
	if err != nil {
		return nil, err
	}

	// Generate Refresh Token
	refreshTokenString, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	// Hash and store refresh token
	refreshToken := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: utils.HashToken(refreshTokenString),
		ExpiresAt: time.Now().Add(s.config.RefreshTokenExpire),
		IsRevoked: false,
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, err
	}

	return &responses.LoginResponse{
		User:         mapper.UserToResponse(user),
		Token:        accessToken,
		RefreshToken: refreshTokenString,
	}, nil
}
