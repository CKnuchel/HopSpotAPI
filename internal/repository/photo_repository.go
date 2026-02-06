package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"hopSpotAPI/internal/domain"
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

func (r *photoRepository) HardDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.Photo{}, id).Error
}

func (r *photoRepository) FindBySpotID(ctx context.Context, spotID uint) ([]domain.Photo, error) {
	var photos []domain.Photo
	if err := r.db.WithContext(ctx).Where("spot_id = ?", spotID).Find(&photos).Error; err != nil {
		return nil, err
	}
	return photos, nil
}

func (r *photoRepository) CountBySpotID(ctx context.Context, spotID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.Photo{}).Where("spot_id = ?", spotID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *photoRepository) SetMainPhoto(ctx context.Context, photoID uint, spotID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unset previous main photo
		if err := tx.Model(&domain.Photo{}).Where("spot_id = ? AND is_main = ?", spotID, true).Update("is_main", false).Error; err != nil {
			return err
		}
		// Set new main photo
		if err := tx.Model(&domain.Photo{}).Where("id = ? AND spot_id = ?", photoID, spotID).Update("is_main", true).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *photoRepository) GetMainPhoto(ctx context.Context, spotID uint) (*domain.Photo, error) {
	var photo domain.Photo
	err := r.db.WithContext(ctx).Where("spot_id = ? AND is_main = ?", spotID, true).First(&photo).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No main photo found
		}
		return nil, err
	}
	return &photo, nil
}
