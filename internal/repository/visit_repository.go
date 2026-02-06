package repository

import (
	"context"

	"gorm.io/gorm"

	"hopSpotAPI/internal/domain"
)

type visitRepository struct {
	db *gorm.DB
}

func NewVisitRepository(db *gorm.DB) VisitRepository {
	return &visitRepository{db: db}
}

func (r *visitRepository) Create(ctx context.Context, visit *domain.Visit) error {
	return r.db.WithContext(ctx).Create(visit).Error
}

func (r *visitRepository) FindByID(ctx context.Context, id uint) (*domain.Visit, error) {
	var visit domain.Visit
	if err := r.db.WithContext(ctx).Preload("Spot").Preload("User").First(&visit, id).Error; err != nil {
		return nil, err
	}
	return &visit, nil
}

func (r *visitRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Visit{}, id).Error
}

func (r *visitRepository) FindByUserID(ctx context.Context, userID uint, filter VisitFilter) ([]domain.Visit, int64, error) {
	var visits []domain.Visit
	var count int64

	query := r.db.WithContext(ctx).Model(&domain.Visit{}).Where("user_id = ?", userID)

	// Apply filters
	if filter.SpotID != nil && *filter.SpotID > 0 {
		query = query.Where("spot_id = ?", filter.SpotID)
	}

	// Count total records
	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Page > 0 && filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset)
	}

	if err := query.Preload("Spot").Find(&visits).Error; err != nil {
		return nil, 0, err
	}

	return visits, count, nil
}

func (r *visitRepository) CountBySpotID(ctx context.Context, spotID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Visit{}).Where("spot_id = ?", spotID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
