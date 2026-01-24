package repository

import (
	"context"
	"hopSpotAPI/internal/domain"

	"gorm.io/gorm"
)

type benchRepository struct {
	db *gorm.DB
}

// NewBenchRepository Constructor for BenchRepository
func NewBenchRepository(db *gorm.DB) BenchRepository {
	return &benchRepository{db: db}
}

func (r benchRepository) Create(ctx context.Context, bench *domain.Bench) error {
	return r.db.WithContext(ctx).Create(bench).Error
}

func (r benchRepository) FindByID(ctx context.Context, id uint) (*domain.Bench, error) {
	var bench domain.Bench
	if err := r.db.WithContext(ctx).First(&bench, id).Error; err != nil {
		return nil, err
	}
	return &bench, nil
}

func (r benchRepository) Update(ctx context.Context, bench *domain.Bench) error {
	return r.db.WithContext(ctx).Save(bench).Error
}

func (r benchRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Bench{}, id).Error
}

func (r benchRepository) FindAll(ctx context.Context, filter BenchFilter) ([]domain.Bench, int64, error) {
	var benches []domain.Bench
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Bench{})

	if filter.HasToilet != nil {
		query = query.Where("has_toilet = ?", *filter.HasToilet)
	}

	if filter.HasTrashBin != nil {
		query = query.Where("has_trash_bin = ?", *filter.HasTrashBin)
	}

	if filter.MinRating != nil {
		query = query.Where("rating >= ?", *filter.MinRating)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset).Limit(filter.Limit)
	}

	// Sorting
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}

	sortOrder := "desc"
	if filter.SortOrder == "asc" {
		sortOrder = "asc"
	}

	query = query.Order(sortBy + " " + sortOrder)

	// Execute query
	if err := query.Find(&benches).Error; err != nil {
		return nil, 0, err
	}

	return benches, total, nil
}

func (r benchRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&domain.Bench{}).Where("id = ?", id).Updates(fields).Error
}
