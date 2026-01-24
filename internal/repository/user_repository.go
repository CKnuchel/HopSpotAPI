package repository

import (
	"context"
	"hopSpotAPI/internal/domain"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository Constructor for UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r userRepository) FindByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r userRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, id).Error
}

func (r userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r userRepository) FindAll(ctx context.Context, filter UserFilter) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.User{})

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	if filter.Role != nil {
		query = query.Where("role = ?", *filter.Role)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("email LIKE ? OR display_name LIKE ?", searchPattern, searchPattern)
	}

	query.Count(&total)

	if filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset).Limit(filter.Limit)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r userRepository) UpdateFCMToken(ctx context.Context, userID uint, token string) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).Update("fcm_token", token).Error
}

func (r userRepository) GetAllFCMTokens(ctx context.Context, excludeUserID uint) ([]string, error) {
	var tokens []string
	if err := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id != ? AND fcm_token IS NOT NULL", excludeUserID).
		Pluck("fcm_token", &tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}
