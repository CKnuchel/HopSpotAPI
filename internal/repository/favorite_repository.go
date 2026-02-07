package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"hopSpotAPI/internal/domain"
)

type favoriteRepository struct {
	db *gorm.DB
}

func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &favoriteRepository{db: db}
}

func (r *favoriteRepository) Create(ctx context.Context, favorite *domain.Favorite) error {
	return r.db.WithContext(ctx).Create(favorite).Error
}

func (r *favoriteRepository) Delete(ctx context.Context, userID, spotID uint) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND spot_id = ?", userID, spotID).
		Delete(&domain.Favorite{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("favorite not found")
	}
	return nil
}

func (r *favoriteRepository) DeleteBySpotID(ctx context.Context, spotID uint) error {
	return r.db.WithContext(ctx).Where("spot_id = ?", spotID).Delete(&domain.Favorite{}).Error
}

func (r *favoriteRepository) Exists(ctx context.Context, userID, spotID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Favorite{}).
		Where("user_id = ? AND spot_id = ?", userID, spotID).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *favoriteRepository) FindByUserID(ctx context.Context, userID uint, filter FavoriteFilter) ([]domain.Favorite, int64, error) {
	var favorites []domain.Favorite
	var total int64

	query := r.db.WithContext(ctx).
		Model(&domain.Favorite{}).
		Where("user_id = ?", userID)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset).Limit(filter.Limit)
	}

	// Execute query with preload
	err := query.
		Preload("Spot").
		Preload("Spot.Creator").
		Order("created_at DESC").
		Find(&favorites).Error

	if err != nil {
		return nil, 0, err
	}

	return favorites, total, nil
}

func (r *favoriteRepository) GetSpotIDsByUserID(ctx context.Context, userID uint) ([]uint, error) {
	var spotIDs []uint
	err := r.db.WithContext(ctx).
		Model(&domain.Favorite{}).
		Where("user_id = ?", userID).
		Pluck("spot_id", &spotIDs).Error

	if err != nil {
		return nil, err
	}
	return spotIDs, nil
}
