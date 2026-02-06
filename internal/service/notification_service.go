package service

import (
	"context"
	"fmt"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/apperror"
	"hopSpotAPI/pkg/notification"
)

type NotificationService interface {
	NotifyNewSpot(ctx context.Context, spot *domain.Spot, creatorID uint) error
}

type notificationService struct {
	fcmClient *notification.FCMClient
	userRepo  repository.UserRepository
}

func NewNotificationService(fcmClient *notification.FCMClient, userRepo repository.UserRepository) NotificationService {
	return &notificationService{
		fcmClient: fcmClient,
		userRepo:  userRepo,
	}
}

// NotifyNewSpot implements NotificationService.
func (s *notificationService) NotifyNewSpot(ctx context.Context, spot *domain.Spot, creatorID uint) error {
	// Skip if FCM not configured
	if s.fcmClient == nil {
		return nil
	}

	tokens, err := s.userRepo.GetAllFCMTokens(ctx, creatorID)
	if err != nil {
		return fmt.Errorf("failed to read FCM tokens: %w", err)
	}

	if len(tokens) == 0 {
		return nil
	}

	user, err := s.userRepo.FindByID(ctx, creatorID)
	if err != nil {
		return fmt.Errorf("failed to get notification creator: %w", err)
	}
	if user == nil {
		return apperror.ErrUserNotFound
	}

	data := map[string]string{
		"spot_id": fmt.Sprintf("%d", spot.ID),
		"type":    "new_spot",
	}

	title := "Neuer HopSpot"
	body := fmt.Sprintf("%s hat einen neuen HopSpot hinzugefügt: %s", user.DisplayName, spot.Name)

	err = s.fcmClient.SendToMultiple(ctx, tokens, title, body, data)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}
