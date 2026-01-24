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

type InvitationRepository interface {
	Create(ctx context.Context, code *domain.InvitationCode) error
	FindByID(ctx context.Context, id uint) (*domain.InvitationCode, error)
	Update(ctx context.Context, code *domain.InvitationCode) error
	Delete(ctx context.Context, id uint) error

	FindByCode(ctx context.Context, code string) (*domain.InvitationCode, error)
	FindAll(ctx context.Context, filter InvitationFilter) ([]domain.InvitationCode, int64, error)
	MarkAsRedeemed(ctx context.Context, codeID uint, userID uint) error
}

type BenchRepository interface {
	Create(ctx context.Context, bench *domain.Bench) error
	FindByID(ctx context.Context, id uint) (*domain.Bench, error)
	Update(ctx context.Context, bench *domain.Bench) error
	Delete(ctx context.Context, id uint) error

	FindAll(ctx context.Context, filter BenchFilter) ([]domain.Bench, int64, error)
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
}

type UserFilter struct {
	Page     int
	Limit    int
	IsActive *bool
	Role     *string
	Search   string // search by email or display name
}

type InvitationFilter struct {
	Page       int
	Limit      int
	IsRedeemed *bool
	CreatedBy  *uint
}

type BenchFilter struct {
	Page        int
	Limit       int
	SortBy      string // name, rating, created_at, distance, visit_count
	SortOrder   string // asc, desc
	HasToilet   *bool
	HasTrashBin *bool
	MinRating   *int
	Search      string

	Lat *float64
	Lon *float64
}
