package service

import (
	"context"
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestAdminService_ListUsers_Success(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	svc := NewAdminService(userRepo, invitationRepo)

	users := []domain.User{
		{
			Model:       &gorm.Model{ID: 1},
			Email:       "user1@example.com",
			DisplayName: "User 1",
			Role:        domain.RoleUser,
			IsActive:    true,
		},
		{
			Model:       &gorm.Model{ID: 2},
			Email:       "admin@example.com",
			DisplayName: "Admin",
			Role:        domain.RoleAdmin,
			IsActive:    true,
		},
	}

	req := &requests.ListUsersRequest{
		Page:  1,
		Limit: 50,
	}

	userRepo.EXPECT().
		FindAll(mock.Anything, mock.AnythingOfType("repository.UserFilter")).
		Return(users, int64(2), nil)

	// Act
	result, err := svc.ListUsers(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Users, 2)
	assert.Equal(t, int64(2), result.Pagination.Total)
	assert.Equal(t, 1, result.Pagination.Page)
}

func TestAdminService_ListUsers_WithFilter(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	svc := NewAdminService(userRepo, invitationRepo)

	users := []domain.User{
		{
			Model:       &gorm.Model{ID: 2},
			Email:       "admin@example.com",
			DisplayName: "Admin",
			Role:        domain.RoleAdmin,
			IsActive:    true,
		},
	}

	isActive := true
	req := &requests.ListUsersRequest{
		Page:     1,
		Limit:    50,
		IsActive: &isActive,
		Role:     "admin",
	}

	userRepo.EXPECT().
		FindAll(mock.Anything, mock.MatchedBy(func(f repository.UserFilter) bool {
			return f.IsActive != nil && *f.IsActive == true &&
				f.Role != nil && *f.Role == "admin"
		})).
		Return(users, int64(1), nil)

	// Act
	result, err := svc.ListUsers(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Users, 1)
	assert.Equal(t, "admin", result.Users[0].Role)
}

func TestAdminService_UpdateUser_Success(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	svc := NewAdminService(userRepo, invitationRepo)

	user := &domain.User{
		Model:       &gorm.Model{ID: 1},
		Email:       "user@example.com",
		DisplayName: "Test User",
		Role:        domain.RoleUser,
		IsActive:    true,
	}

	newRole := domain.RoleAdmin
	isActive := false
	req := &requests.AdminUpdateUserRequest{
		Role:     &newRole,
		IsActive: &isActive,
	}

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(user, nil)

	userRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.User")).
		Run(func(ctx context.Context, u *domain.User) {
			assert.Equal(t, domain.RoleAdmin, u.Role)
			assert.False(t, u.IsActive)
		}).
		Return(nil)

	// Act
	result, err := svc.UpdateUser(context.Background(), uint(1), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "admin", result.Role)
}

func TestAdminService_UpdateUser_UserNotFound(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	svc := NewAdminService(userRepo, invitationRepo)

	newRole := domain.RoleAdmin
	req := &requests.AdminUpdateUserRequest{
		Role: &newRole,
	}

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(999)).
		Return(nil, nil)

	// Act
	result, err := svc.UpdateUser(context.Background(), uint(999), req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user not found")
}

func TestAdminService_UpdateUser_PartialUpdate(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	svc := NewAdminService(userRepo, invitationRepo)

	user := &domain.User{
		Model:       &gorm.Model{ID: 1},
		Email:       "user@example.com",
		DisplayName: "Test User",
		Role:        domain.RoleUser,
		IsActive:    true,
	}

	// Only update isActive, not role
	isActive := false
	req := &requests.AdminUpdateUserRequest{
		IsActive: &isActive,
	}

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(1)).
		Return(user, nil)

	userRepo.EXPECT().
		Update(mock.Anything, mock.AnythingOfType("*domain.User")).
		Run(func(ctx context.Context, u *domain.User) {
			// Role should remain unchanged
			assert.Equal(t, domain.RoleUser, u.Role)
			assert.False(t, u.IsActive)
		}).
		Return(nil)

	// Act
	result, err := svc.UpdateUser(context.Background(), uint(1), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "user", result.Role) // Unchanged
}

func TestAdminService_DeleteUser_Success(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	svc := NewAdminService(userRepo, invitationRepo)

	user := &domain.User{
		Model:       &gorm.Model{ID: 2},
		Email:       "user@example.com",
		DisplayName: "Test User",
	}

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(2)).
		Return(user, nil)

	userRepo.EXPECT().
		Delete(mock.Anything, uint(2)).
		Return(nil)

	// Act - admin (ID 1) deletes user (ID 2)
	err := svc.DeleteUser(context.Background(), uint(2), uint(1))

	// Assert
	assert.NoError(t, err)
}

func TestAdminService_DeleteUser_CannotDeleteSelf(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	svc := NewAdminService(userRepo, invitationRepo)

	// Act - admin tries to delete themselves
	err := svc.DeleteUser(context.Background(), uint(1), uint(1))

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete")
}

func TestAdminService_DeleteUser_UserNotFound(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	svc := NewAdminService(userRepo, invitationRepo)

	userRepo.EXPECT().
		FindByID(mock.Anything, uint(999)).
		Return(nil, nil)

	// Act
	err := svc.DeleteUser(context.Background(), uint(999), uint(1))

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestAdminService_ListInvitationCodes_Success(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	svc := NewAdminService(userRepo, invitationRepo)

	adminID := uint(1)
	adminUser := domain.User{
		Model:       &gorm.Model{ID: 1},
		Email:       "admin@example.com",
		DisplayName: "Admin",
		Role:        domain.RoleAdmin,
	}

	codes := []domain.InvitationCode{
		{
			Model:      &gorm.Model{ID: 1},
			Code:       "ABC123",
			Comment:    "For friend",
			CreatedBy:  &adminID,
			Creator:    &adminUser,
			RedeemedBy: nil,
		},
		{
			Model:      &gorm.Model{ID: 2},
			Code:       "XYZ789",
			Comment:    "For colleague",
			CreatedBy:  &adminID,
			Creator:    &adminUser,
			RedeemedBy: nil,
		},
	}

	req := &requests.ListInvitationCodesRequest{
		Page:  1,
		Limit: 50,
	}

	invitationRepo.EXPECT().
		FindAll(mock.Anything, mock.AnythingOfType("repository.InvitationFilter")).
		Return(codes, int64(2), nil)

	// Act
	result, err := svc.ListInvitationCodes(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Codes, 2)
	assert.Equal(t, int64(2), result.Pagination.Total)
}

func TestAdminService_ListInvitationCodes_FilterRedeemed(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	svc := NewAdminService(userRepo, invitationRepo)

	isRedeemed := true
	req := &requests.ListInvitationCodesRequest{
		Page:       1,
		Limit:      50,
		IsRedeemed: &isRedeemed,
	}

	invitationRepo.EXPECT().
		FindAll(mock.Anything, mock.MatchedBy(func(f repository.InvitationFilter) bool {
			return f.IsRedeemed != nil && *f.IsRedeemed == true
		})).
		Return([]domain.InvitationCode{}, int64(0), nil)

	// Act
	result, err := svc.ListInvitationCodes(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Codes, 0)
}

func TestAdminService_CreateInvitationCode_Success(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	svc := NewAdminService(userRepo, invitationRepo)

	req := &requests.CreateInvitationCodeRequest{
		Comment: "For new team member",
	}

	invitationRepo.EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.InvitationCode")).
		Run(func(ctx context.Context, code *domain.InvitationCode) {
			// Verify code was generated
			assert.NotEmpty(t, code.Code)
			assert.Len(t, code.Code, 6)
			assert.Equal(t, "For new team member", code.Comment)
			assert.Equal(t, uint(1), *code.CreatedBy)
			// Simulate DB setting ID
			code.Model = &gorm.Model{ID: 1}
			// Simulate DB preloading Creator
			code.Creator = &domain.User{
				Model:       &gorm.Model{ID: 1},
				Email:       "admin@example.com",
				DisplayName: "Admin",
				Role:        domain.RoleAdmin,
			}
		}).
		Return(nil)

	// Act
	result, err := svc.CreateInvitationCode(context.Background(), req, uint(1))

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.Code)
	assert.Equal(t, "For new team member", result.Comment)
}

func TestAdminService_ListUsers_Pagination(t *testing.T) {
	// Arrange
	userRepo := mocks.NewUserRepository(t)
	invitationRepo := mocks.NewInvitationRepository(t)
	svc := NewAdminService(userRepo, invitationRepo)

	// 55 total users, requesting page 2 with limit 50
	users := make([]domain.User, 5) // Only 5 users on page 2
	for i := 0; i < 5; i++ {
		users[i] = domain.User{
			Model:       &gorm.Model{ID: uint(51 + i)},
			Email:       "user@example.com",
			DisplayName: "User",
			Role:        domain.RoleUser,
		}
	}

	req := &requests.ListUsersRequest{
		Page:  2,
		Limit: 50,
	}

	userRepo.EXPECT().
		FindAll(mock.Anything, mock.MatchedBy(func(f repository.UserFilter) bool {
			return f.Page == 2 && f.Limit == 50
		})).
		Return(users, int64(55), nil)

	// Act
	result, err := svc.ListUsers(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Users, 5)
	assert.Equal(t, int64(55), result.Pagination.Total)
	assert.Equal(t, 2, result.Pagination.TotalPages) // 55/50 = 2 pages
	assert.Equal(t, 2, result.Pagination.Page)
}
