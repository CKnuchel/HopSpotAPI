package repository

import (
	"context"

	"gorm.io/gorm"

	"hopSpotAPI/internal/domain"
)

type activityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) Create(ctx context.Context, activity *domain.Activity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

func (r *activityRepository) DeleteBySpotID(ctx context.Context, spotID uint) error {
	return r.db.WithContext(ctx).Where("spot_id = ?", spotID).Delete(&domain.Activity{}).Error
}

func (r *activityRepository) FindAll(ctx context.Context, filter ActivityFilter) ([]domain.Activity, int64, error) {
	var activities []domain.Activity
	var count int64

	query := r.db.WithContext(ctx).Model(&domain.Activity{})

	// Apply action_type filter
	if filter.ActionType != nil && *filter.ActionType != "" {
		query = query.Where("action_type = ?", *filter.ActionType)
	}

	// Count total records
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Order by created_at descending (newest first)
	query = query.Order("created_at DESC")

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Page > 0 && filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset)
	}

	// Load relations
	if err := query.Preload("User").Preload("Spot").Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, count, nil
}
