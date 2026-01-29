package repository

import (
	"context"
	"errors"
	"hopSpotAPI/internal/domain"

	"gorm.io/gorm"
)

type photoRepository struct {
	db *gorm.DB
}

func NewPhotoRepository(db *gorm.DB) PhotoRepository {
	return &photoRepository{db: db}
}

func (r *photoRepository) Create(ctx context.Context, photo *domain.Photo) error {
	return r.db.WithContext(ctx).Create(photo).Error
}

func (r *photoRepository) FindByID(ctx context.Context, id uint) (*domain.Photo, error) {
	var photo domain.Photo
	err := r.db.WithContext(ctx).First(&photo, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &photo, nil
}

func (r *photoRepository) Update(ctx context.Context, photo *domain.Photo) error {
	return r.db.WithContext(ctx).Save(photo).Error
}

func (r *photoRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Photo{}, id).Error
}

func (r *photoRepository) FindByBenchID(ctx context.Context, benchID uint) ([]domain.Photo, error) {
	var photos []domain.Photo
	if err := r.db.WithContext(ctx).Where("bench_id = ?", benchID).Find(&photos).Error; err != nil {
		return nil, err
	}
	return photos, nil
}

func (r *photoRepository) CountByBenchID(ctx context.Context, benchID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Photo{}).Where("bench_id = ?", benchID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *photoRepository) SetMainPhoto(ctx context.Context, photoID uint, benchID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unset previous main photo
		if err := tx.Model(&domain.Photo{}).Where("bench_id = ? AND is_main = ?", benchID, true).Update("is_main", false).Error; err != nil {
			return err
		}
		// Set new main photo
		if err := tx.Model(&domain.Photo{}).Where("id = ? AND bench_id = ?", photoID, benchID).Update("is_main", true).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *photoRepository) GetMainPhoto(ctx context.Context, benchID uint) (*domain.Photo, error) {
	var photo domain.Photo
	err := r.db.WithContext(ctx).Where("bench_id = ? AND is_main = ?", benchID, true).First(&photo).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No main photo found
		}
		return nil, err
	}
	return &photo, nil
}
