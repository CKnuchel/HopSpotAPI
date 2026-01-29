package service

import (
	"context"
	"hopSpotAPI/internal/config"
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/mocks"
	"hopSpotAPI/pkg/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestUserService_GetProfile_Success(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	cfg := config.Config{}
	svc := NewUserService(userRepo, cfg)

	user := &domain.User{
		Model:       &gorm.Model{ID: 1},
		Email:       "test@example.com",
		DisplayName: "Test User",
		Role:        domain.RoleUser,
		IsActive:    true,
	}

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(user, nil)

	// Act
	result, err := svc.GetProfile(context.Background(), uint(1))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "Test User", result.DisplayName)
	assert.Equal(t, "user", result.Role)
}

func TestUserService_GetProfile_UserNotFound(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	cfg := config.Config{}
	svc := NewUserService(userRepo, cfg)

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(999)).
		Return(nil, nil) // User not found

	// Act
	result, err := svc.GetProfile(context.Background(), uint(999))

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserService_UpdateProfile_Success(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	cfg := config.Config{}
	svc := NewUserService(userRepo, cfg)

	user := &domain.User{
		Model:       &gorm.Model{ID: 1},
		Email:       "test@example.com",
		DisplayName: "Old Name",
		Role:        domain.RoleUser,
		IsActive:    true,
	}

	newName := "New Name"
	req := &requests.UpdateProfileRequest{
		DisplayName: &newName,
	}

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(user, nil)

	userRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.User")).
		Run(func(ctx context.Context, u *domain.User) {
			assert.Equal(t, "New Name", u.DisplayName)
		}).
		Return(nil)

	// Act
	result, err := svc.UpdateProfile(context.Background(), uint(1), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New Name", result.DisplayName)
}

func TestUserService_UpdateProfile_UserNotFound(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	cfg := config.Config{}
	svc := NewUserService(userRepo, cfg)

	newName := "New Name"
	req := &requests.UpdateProfileRequest{
		DisplayName: &newName,
	}

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(999)).
		Return(nil, nil) // User not found

	// Act
	result, err := svc.UpdateProfile(context.Background(), uint(999), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserService_UpdateProfile_NoChanges(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	cfg := config.Config{}
	svc := NewUserService(userRepo, cfg)

	user := &domain.User{
		Model:       &gorm.Model{ID: 1},
		Email:       "test@example.com",
		DisplayName: "Original Name",
		Role:        domain.RoleUser,
		IsActive:    true,
	}

	// Empty request - no changes
	req := &requests.UpdateProfileRequest{}

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(user, nil)

	userRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.User")).
		Return(nil)

	// Act
	result, err := svc.UpdateProfile(context.Background(), uint(1), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Original Name", result.DisplayName) // Unchanged
}

func TestUserService_ChangePassword_Success(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	cfg := config.Config{}
	svc := NewUserService(userRepo, cfg)

	oldPassword := "OldPassword123!"
	hashedOldPassword, _ := utils.HashPassword(oldPassword)

	user := &domain.User{
		Model:        &gorm.Model{ID: 1},
		Email:        "test@example.com",
		PasswordHash: hashedOldPassword,
		DisplayName:  "Test User",
		Role:         domain.RoleUser,
		IsActive:     true,
	}

	req := &requests.ChangePasswordRequest{
		OldPassword: oldPassword,
		NewPassword: "NewPassword456!",
	}

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(user, nil)

	userRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.User")).
		Run(func(ctx context.Context, u *domain.User) {
			// Verify password was changed (hash should be different)
			assert.NotEqual(t, hashedOldPassword, u.PasswordHash)
			// Verify new password is valid
			assert.True(t, utils.CheckPasswordHash("NewPassword456!", u.PasswordHash))
		}).
		Return(nil)

	// Act
	err := svc.ChangePassword(context.Background(), uint(1), req)

	// Assert
	assert.NoError(t, err)
}

func TestUserService_ChangePassword_UserNotFound(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	cfg := config.Config{}
	svc := NewUserService(userRepo, cfg)

	req := &requests.ChangePasswordRequest{
		OldPassword: "OldPassword123!",
		NewPassword: "NewPassword456!",
	}

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(999)).
		Return(nil, nil) // User not found

	// Act
	err := svc.ChangePassword(context.Background(), uint(999), req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserService_ChangePassword_WrongOldPassword(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	cfg := config.Config{}
	svc := NewUserService(userRepo, cfg)

	hashedPassword, _ := utils.HashPassword("CorrectPassword123!")

	user := &domain.User{
		Model:        &gorm.Model{ID: 1},
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		DisplayName:  "Test User",
		Role:         domain.RoleUser,
		IsActive:     true,
	}

	req := &requests.ChangePasswordRequest{
		OldPassword: "WrongPassword123!", // Wrong password
		NewPassword: "NewPassword456!",
	}

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(user, nil)

	// Act
	err := svc.ChangePassword(context.Background(), uint(1), req)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}
