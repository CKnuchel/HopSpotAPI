package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"hopSpotAPI/internal/domain"
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
	err := r.db.WithContext(ctx).
		Preload("Creator").
		First(&bench, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
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

	// Apply filters
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
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	// Radius filter (requires lat/lon)
	if filter.Radius != nil && filter.Lat != nil && filter.Lon != nil {
		// Haversine formula for radius filtering (in meters)
		radiusKm := float64(*filter.Radius) / 1000.0
		haversineCondition := fmt.Sprintf(
			"(6371 * acos(cos(radians(%f)) * cos(radians(latitude)) * cos(radians(longitude) - radians(%f)) + sin(radians(%f)) * sin(radians(latitude)))) <= %f",
			*filter.Lat, *filter.Lon, *filter.Lat, radiusKm,
		)
		query = query.Where(haversineCondition)
	}

	// Count total records (before pagination)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Handle distance sorting
	if filter.SortBy == "distance" {
		if filter.Lat == nil || filter.Lon == nil {
			return nil, 0, errors.New("lat and lon are required for distance sorting")
		}

		// Haversine formula for distance calculation (result in km)
		distanceFormula := fmt.Sprintf(
			"(6371 * acos(LEAST(1.0, cos(radians(%f)) * cos(radians(latitude)) * cos(radians(longitude) - radians(%f)) + sin(radians(%f)) * sin(radians(latitude)))))",
			*filter.Lat, *filter.Lon, *filter.Lat,
		)

		sortOrder := "ASC"
		if filter.SortOrder == "desc" {
			sortOrder = "DESC"
		}

		query = query.Order(fmt.Sprintf("%s %s", distanceFormula, sortOrder))
	} else {
		// Normal column sorting
		sortBy := "created_at"
		validSortColumns := map[string]bool{
			"name":       true,
			"rating":     true,
			"created_at": true,
			"updated_at": true,
		}

		if filter.SortBy != "" && validSortColumns[filter.SortBy] {
			sortBy = filter.SortBy
		}

		sortOrder := "DESC"
		if filter.SortOrder == "asc" {
			sortOrder = "ASC"
		}

		query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
	}

	// Apply pagination
	if filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset).Limit(filter.Limit)
	}

	// Execute query
	if err := query.Find(&benches).Error; err != nil {
		return nil, 0, err
	}

	return benches, total, nil
}

func (r benchRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&domain.Bench{}).Where("id = ?", id).Updates(fields).Error
}
