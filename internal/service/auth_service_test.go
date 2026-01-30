package service

import (
	"context"
	"testing"
	"time"

	"hopSpotAPI/internal/config"
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/mocks"
	"hopSpotAPI/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestAuthService_Register_Success(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	refreshTokenRepo := mocks.NewRefreshTokenRepository(t)

	cfg := config.Config{
		JWTSecret:          "test-secret-min-32-characters-long",
		JWTExpire:          3600 * time.Second,
		RefreshTokenExpire: 90 * 24 * time.Hour,
		JWTIssuer:          "test",
		JWTAudience:        "test",
	}

	svc := NewAuthService(userRepo, invitationRepo, refreshTokenRepo, cfg)

	req := &requests.RegisterRequest{
		Email:          "test@example.com",
		Password:       "TestPass123!",
		DisplayName:    "Test User",
		InvitationCode: "ABC12345",
	}

	// Mock expectations - IN ORDER!

	// 1. Check if email exists (should return nil, nil = not found)
	userRepo.EXPECT().
		FindByEmail(mock.Anything, "test@example.com").
		Return(nil, nil)

	// 2. Check invitation code
	invitationRepo.EXPECT().
		FindByCode(mock.Anything, "ABC12345").
		Return(&domain.InvitationCode{
			Model:      &gorm.Model{ID: 1},
			Code:       "ABC12345",
			RedeemedBy: nil,
		}, nil)

	// 3. Count users (0 = first user becomes admin)
	userRepo.EXPECT().
		Count(mock.Anything).
		Return(int64(0), nil)

	// 4. Create user - IMPORTANT: Set the ID via Run()
	userRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.User")).
		Run(func(ctx context.Context, user *domain.User) {
			// Simulate what GORM does: set the Model with ID
			user.Model = &gorm.Model{ID: 1}
		}).
		Return(nil)

	// 5. Mark invitation as redeemed
	invitationRepo.EXPECT().
		MarkAsRedeemed(mock.Anything, uint(1), uint(1)).
		Return(nil)

	// 6. Create refresh token (in generateTokens)
	refreshTokenRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).
		Return(nil)

	// Act
	result, err := svc.Register(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Token)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, "test@example.com", result.User.Email)
	assert.Equal(t, "admin", result.User.Role) // First user becomes admin
}

func TestAuthService_Register_SecondUserBecomesUser(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	refreshTokenRepo := mocks.NewRefreshTokenRepository(t)

	cfg := config.Config{
		JWTSecret:          "test-secret-min-32-characters-long",
		JWTExpire:          3600 * time.Second,
		RefreshTokenExpire: 90 * 24 * time.Hour,
		JWTIssuer:          "test",
		JWTAudience:        "test",
	}

	svc := NewAuthService(userRepo, invitationRepo, refreshTokenRepo, cfg)

	req := &requests.RegisterRequest{
		Email:          "second@example.com",
		Password:       "TestPass123!",
		DisplayName:    "Second User",
		InvitationCode: "XYZ98765",
	}

	// Mock expectations
	userRepo.EXPECT().
		FindByEmail(mock.Anything, "second@example.com").
		Return(nil, nil)

	invitationRepo.EXPECT().
		FindByCode(mock.Anything, "XYZ98765").
		Return(&domain.InvitationCode{
			Model:      &gorm.Model{ID: 2},
			Code:       "XYZ98765",
			RedeemedBy: nil,
		}, nil)

	// Count returns 1 = not first user
	userRepo.EXPECT().
		Count(mock.Anything).
		Return(int64(1), nil)

	userRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.User")).
		Run(func(ctx context.Context, user *domain.User) {
			user.Model = &gorm.Model{ID: 2}
		}).
		Return(nil)

	invitationRepo.EXPECT().
		MarkAsRedeemed(mock.Anything, uint(2), uint(2)).
		Return(nil)

	refreshTokenRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).
		Return(nil)

	// Act
	result, err := svc.Register(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user", result.User.Role) // Second user becomes regular user
}

func TestAuthService_Register_InvalidInvitationCode(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	refreshTokenRepo := mocks.NewRefreshTokenRepository(t)

	cfg := config.Config{
		JWTSecret:          "test-secret-min-32-characters-long",
		JWTExpire:          3600 * time.Second,
		RefreshTokenExpire: 90 * 24 * time.Hour,
	}

	svc := NewAuthService(userRepo, invitationRepo, refreshTokenRepo, cfg)

	req := &requests.RegisterRequest{
		Email:          "test@example.com",
		Password:       "TestPass123!",
		DisplayName:    "Test User",
		InvitationCode: "INVALID",
	}

	// Mock expectations
	userRepo.EXPECT().
		FindByEmail(mock.Anything, "test@example.com").
		Return(nil, nil)

	invitationRepo.EXPECT().
		FindByCode(mock.Anything, "INVALID").
		Return(nil, nil) // Not found = nil, nil

	// Act
	result, err := svc.Register(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid invitation code")
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	refreshTokenRepo := mocks.NewRefreshTokenRepository(t)

	cfg := config.Config{
		JWTSecret:          "test-secret-min-32-characters-long",
		JWTExpire:          3600 * time.Second,
		RefreshTokenExpire: 90 * 24 * time.Hour,
	}

	svc := NewAuthService(userRepo, invitationRepo, refreshTokenRepo, cfg)

	req := &requests.RegisterRequest{
		Email:          "existing@example.com",
		Password:       "TestPass123!",
		DisplayName:    "Test User",
		InvitationCode: "ABC12345",
	}

	existingUser := &domain.User{
		Model: &gorm.Model{ID: 1},
		Email: "existing@example.com",
	}

	// Mock expectations - email already exists
	userRepo.EXPECT().
		FindByEmail(mock.Anything, "existing@example.com").
		Return(existingUser, nil)

	// Act
	result, err := svc.Register(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "email already exists")
}

func TestAuthService_Register_InvitationCodeAlreadyRedeemed(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	refreshTokenRepo := mocks.NewRefreshTokenRepository(t)

	cfg := config.Config{
		JWTSecret:          "test-secret-min-32-characters-long",
		JWTExpire:          3600 * time.Second,
		RefreshTokenExpire: 90 * 24 * time.Hour,
	}

	svc := NewAuthService(userRepo, invitationRepo, refreshTokenRepo, cfg)

	req := &requests.RegisterRequest{
		Email:          "test@example.com",
		Password:       "TestPass123!",
		DisplayName:    "Test User",
		InvitationCode: "USED123",
	}

	redeemedByID := uint(99)

	userRepo.EXPECT().
		FindByEmail(mock.Anything, "test@example.com").
		Return(nil, nil)

	invitationRepo.EXPECT().
		FindByCode(mock.Anything, "USED123").
		Return(&domain.InvitationCode{
			Model:      &gorm.Model{ID: 1},
			Code:       "USED123",
			RedeemedBy: &redeemedByID, // Already redeemed!
		}, nil)

	// Act
	result, err := svc.Register(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "already redeemed")
}

func TestAuthService_Login_Success(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	refreshTokenRepo := mocks.NewRefreshTokenRepository(t)

	cfg := config.Config{
		JWTSecret:          "test-secret-min-32-characters-long",
		JWTExpire:          3600 * time.Second,
		RefreshTokenExpire: 90 * 24 * time.Hour,
		JWTIssuer:          "test",
		JWTAudience:        "test",
	}

	svc := NewAuthService(userRepo, invitationRepo, refreshTokenRepo, cfg)

	hashedPassword, _ := utils.HashPassword("TestPass123!")
	user := &domain.User{
		Model:        &gorm.Model{ID: 1},
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		DisplayName:  "Test User",
		Role:         domain.RoleUser,
		IsActive:     true,
	}

	req := &requests.LoginRequest{
		Email:    "test@example.com",
		Password: "TestPass123!",
	}

	// Mock expectations
	userRepo.EXPECT().
		FindByEmail(mock.Anything, "test@example.com").
		Return(user, nil)

	refreshTokenRepo.EXPECT().
		RevokeByUserID(mock.Anything, uint(1)).
		Return(nil)

	refreshTokenRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).
		Return(nil)

	// Act
	result, err := svc.Login(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Token)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, "test@example.com", result.User.Email)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	refreshTokenRepo := mocks.NewRefreshTokenRepository(t)

	cfg := config.Config{
		JWTSecret:          "test-secret-min-32-characters-long",
		JWTExpire:          3600 * time.Second,
		RefreshTokenExpire: 90 * 24 * time.Hour,
	}

	svc := NewAuthService(userRepo, invitationRepo, refreshTokenRepo, cfg)

	req := &requests.LoginRequest{
		Email:    "notfound@example.com",
		Password: "TestPass123!",
	}

	// Mock expectations - user not found
	userRepo.EXPECT().
		FindByEmail(mock.Anything, "notfound@example.com").
		Return(nil, nil)

	// Act
	result, err := svc.Login(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	refreshTokenRepo := mocks.NewRefreshTokenRepository(t)

	cfg := config.Config{
		JWTSecret:          "test-secret-min-32-characters-long",
		JWTExpire:          3600 * time.Second,
		RefreshTokenExpire: 90 * 24 * time.Hour,
	}

	svc := NewAuthService(userRepo, invitationRepo, refreshTokenRepo, cfg)

	hashedPassword, _ := utils.HashPassword("CorrectPassword!")
	user := &domain.User{
		Model:        &gorm.Model{ID: 1},
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		DisplayName:  "Test User",
		Role:         domain.RoleUser,
		IsActive:     true,
	}

	req := &requests.LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword!",
	}

	// Mock expectations
	userRepo.EXPECT().
		FindByEmail(mock.Anything, "test@example.com").
		Return(user, nil)

	// Act
	result, err := svc.Login(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	refreshTokenRepo := mocks.NewRefreshTokenRepository(t)

	cfg := config.Config{
		JWTSecret:          "test-secret-min-32-characters-long",
		JWTExpire:          3600 * time.Second,
		RefreshTokenExpire: 90 * 24 * time.Hour,
	}

	svc := NewAuthService(userRepo, invitationRepo, refreshTokenRepo, cfg)

	hashedPassword, _ := utils.HashPassword("TestPass123!")
	user := &domain.User{
		Model:        &gorm.Model{ID: 1},
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		DisplayName:  "Test User",
		Role:         domain.RoleUser,
		IsActive:     false, // Inactive!
	}

	req := &requests.LoginRequest{
		Email:    "test@example.com",
		Password: "TestPass123!",
	}

	// Mock expectations
	userRepo.EXPECT().
		FindByEmail(mock.Anything, "test@example.com").
		Return(user, nil)

	// Act
	result, err := svc.Login(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "deactivated")
}

func TestAuthService_Logout_Success(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	refreshTokenRepo := mocks.NewRefreshTokenRepository(t)

	cfg := config.Config{
		JWTSecret:          "test-secret-min-32-characters-long",
		JWTExpire:          3600 * time.Second,
		RefreshTokenExpire: 90 * 24 * time.Hour,
	}

	svc := NewAuthService(userRepo, invitationRepo, refreshTokenRepo, cfg)

	refreshToken := "some-refresh-token"
	tokenHash := utils.HashToken(refreshToken)

	req := &requests.LogoutRequest{
		RefreshToken: refreshToken,
	}

	// Mock expectations
	refreshTokenRepo.EXPECT().
		FindByTokenHash(mock.Anything, tokenHash).
		Return(&domain.RefreshToken{
			Model:     &gorm.Model{ID: 1},
			UserID:    1,
			TokenHash: tokenHash,
			IsRevoked: false,
		}, nil)

	refreshTokenRepo.EXPECT().
		RevokeByID(mock.Anything, uint(1)).
		Return(nil)

	// Act
	err := svc.Logout(context.Background(), req)

	// Assert
	assert.NoError(t, err)
}

func TestAuthService_Logout_TokenNotFound(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	refreshTokenRepo := mocks.NewRefreshTokenRepository(t)

	cfg := config.Config{
		JWTSecret:          "test-secret-min-32-characters-long",
		JWTExpire:          3600 * time.Second,
		RefreshTokenExpire: 90 * 24 * time.Hour,
	}

	svc := NewAuthService(userRepo, invitationRepo, refreshTokenRepo, cfg)

	refreshToken := "non-existent-token"
	tokenHash := utils.HashToken(refreshToken)

	req := &requests.LogoutRequest{
		RefreshToken: refreshToken,
	}

	// Mock expectations - token not found
	refreshTokenRepo.EXPECT().
		FindByTokenHash(mock.Anything, tokenHash).
		Return(nil, nil)

	// Act
	err := svc.Logout(context.Background(), req)

	// Assert
	assert.NoError(t, err) // Logout should succeed even if token doesn't exist
}
