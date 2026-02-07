package repository

import (
	"context"

	"gorm.io/gorm"

	"hopSpotAPI/internal/domain"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// FindByRelatedSpotIDUnscoped returns all notifications for a related spot, including soft-deleted ones
func (r *notificationRepository) FindByRelatedSpotIDUnscoped(ctx context.Context, relatedSpotID uint) ([]domain.Notification, error) {
	var notifications []domain.Notification
	if err := r.db.WithContext(ctx).Unscoped().Where("related_spot_id = ?", relatedSpotID).Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) HardDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.Notification{}, id).Error
}
