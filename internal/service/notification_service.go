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
	NotifyNewBench(ctx context.Context, bench *domain.Bench, creatorID uint) error
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

// NotifyNewBench implements NotificationService.
func (s *notificationService) NotifyNewBench(ctx context.Context, bench *domain.Bench, creatorID uint) error {
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
		"bench_id": fmt.Sprintf("%d", bench.ID),
		"type":     "new_bench",
	}

	title := "Neue Bank 🎉"
	body := fmt.Sprintf("%s hat eine neue Bank hinzugefügt: %s", user.DisplayName, bench.Name)

	err = s.fcmClient.SendToMultiple(ctx, tokens, title, body, data)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}
