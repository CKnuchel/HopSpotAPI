package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"hopSpotAPI/internal/domain"
)

type spotRepository struct {
	db *gorm.DB
}

// NewSpotRepository Constructor for SpotRepository
func NewSpotRepository(db *gorm.DB) SpotRepository {
	return &spotRepository{db: db}
}

func (r spotRepository) Create(ctx context.Context, spot *domain.Spot) error {
	return r.db.WithContext(ctx).Create(spot).Error
}

func (r spotRepository) FindByID(ctx context.Context, id uint) (*domain.Spot, error) {
	var spot domain.Spot
	err := r.db.WithContext(ctx).
		Preload("Creator").
		First(&spot, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &spot, nil
}

func (r spotRepository) Update(ctx context.Context, spot *domain.Spot) error {
	return r.db.WithContext(ctx).Save(spot).Error
}

func (r spotRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Spot{}, id).Error
}

func (r spotRepository) FindAll(ctx context.Context, filter SpotFilter) ([]domain.Spot, int64, error) {
	var spots []domain.Spot
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Spot{})

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
	if err := query.Find(&spots).Error; err != nil {
		return nil, 0, err
	}

	return spots, total, nil
}

func (r spotRepository) FindRandom(ctx context.Context) (*domain.Spot, error) {
	var spot domain.Spot
	err := r.db.WithContext(ctx).
		Preload("Creator").
		Order("RANDOM()").
		First(&spot).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &spot, nil
}

func (r spotRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&domain.Spot{}).Where("id = ?", id).Updates(fields).Error
}
