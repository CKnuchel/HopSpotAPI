package repository

import (
	"context"
	"hopSpotAPI/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uint) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uint) error

	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindAll(ctx context.Context, filter UserFilter) ([]domain.User, int64, error)
	UpdateFCMToken(ctx context.Context, userID uint, token string) error
	GetAllFCMTokens(ctx context.Context, excludeUserID uint) ([]string, error)
}

type UserFilter struct {
	Page     int
	Limit    int
	IsActive *bool
	Role     *string
	Search   string // search by email or display name
}
