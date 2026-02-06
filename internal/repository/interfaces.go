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
	Count(ctx context.Context) (int64, error)

	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindAll(ctx context.Context, filter UserFilter) ([]domain.User, int64, error)
	UpdateFCMToken(ctx context.Context, userID uint, token string) error
	GetAllFCMTokens(ctx context.Context, excludeUserID uint) ([]string, error)
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	RevokeByUserID(ctx context.Context, userID uint) error
	RevokeByID(ctx context.Context, id uint) error
	DeleteExpired(ctx context.Context) error
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
	FindRandom(ctx context.Context) (*domain.Bench, error)
	UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error
}

type PhotoRepository interface {
	Create(ctx context.Context, photo *domain.Photo) error
	FindByID(ctx context.Context, id uint) (*domain.Photo, error)
	Update(ctx context.Context, photo *domain.Photo) error
	Delete(ctx context.Context, id uint) error

	FindByBenchID(ctx context.Context, benchID uint) ([]domain.Photo, error)
	CountByBenchID(ctx context.Context, benchID uint) (int64, error)
	SetMainPhoto(ctx context.Context, photoID uint, benchID uint) error
	GetMainPhoto(ctx context.Context, benchID uint) (*domain.Photo, error)
}

type VisitRepository interface {
	Create(ctx context.Context, visit *domain.Visit) error
	FindByID(ctx context.Context, id uint) (*domain.Visit, error)
	Delete(ctx context.Context, id uint) error

	FindByUserID(ctx context.Context, userID uint, filter VisitFilter) ([]domain.Visit, int64, error)
	CountByBenchID(ctx context.Context, benchID uint) (int64, error)
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

	Lat    *float64
	Lon    *float64
	Radius *int // in Metern
}

type VisitFilter struct {
	Page      int
	Limit     int
	SortOrder string // asc, desc
	BenchID   *uint
}

type FavoriteRepository interface {
	Create(ctx context.Context, favorite *domain.Favorite) error
	Delete(ctx context.Context, userID, benchID uint) error
	Exists(ctx context.Context, userID, benchID uint) (bool, error)
	FindByUserID(ctx context.Context, userID uint, filter FavoriteFilter) ([]domain.Favorite, int64, error)
	GetBenchIDsByUserID(ctx context.Context, userID uint) ([]uint, error)
}

type FavoriteFilter struct {
	Page  int
	Limit int
}

type ActivityRepository interface {
	Create(ctx context.Context, activity *domain.Activity) error
	FindAll(ctx context.Context, filter ActivityFilter) ([]domain.Activity, int64, error)
}

type ActivityFilter struct {
	Page       int
	Limit      int
	ActionType *string
}
