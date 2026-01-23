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
	//TODO https://gorm.io/docs/query.html#Retrieving-objects-with-primary-key
	panic("implement me")
}
