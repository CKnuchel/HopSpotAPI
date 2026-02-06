package service

import (
	"context"
	"time"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/storage"
)

type ActivityService interface {
	Create(ctx context.Context, userID uint, actionType string, benchID *uint) error
	List(ctx context.Context, req *requests.ListActivitiesRequest) (*responses.PaginatedActivitiesResponse, error)
}

type activityService struct {
	activityRepo repository.ActivityRepository
	photoRepo    repository.PhotoRepository
	minioClient  *storage.MinioClient
}

func NewActivityService(activityRepo repository.ActivityRepository, photoRepo repository.PhotoRepository, minioClient *storage.MinioClient) ActivityService {
	return &activityService{
		activityRepo: activityRepo,
		photoRepo:    photoRepo,
		minioClient:  minioClient,
	}
}

func (s *activityService) Create(ctx context.Context, userID uint, actionType string, benchID *uint) error {
	activity := &domain.Activity{
		UserID:     userID,
		ActionType: actionType,
		BenchID:    benchID,
		CreatedAt:  time.Now(),
	}

	return s.activityRepo.Create(ctx, activity)
}

func (s *activityService) List(ctx context.Context, req *requests.ListActivitiesRequest) (*responses.PaginatedActivitiesResponse, error) {
	filter := repository.ActivityFilter{
		Page:       req.Page,
		Limit:      req.Limit,
		ActionType: req.ActionType,
	}

	activities, total, err := s.activityRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Calculate pagination info
	totalPages := int(total) / req.Limit
	if int(total)%req.Limit > 0 {
		totalPages++
	}

	// Map activities to responses with photo URLs
	activityResponses := make([]responses.ActivityResponse, len(activities))
	for i, activity := range activities {
		activityResponses[i] = mapper.ActivityToResponse(&activity)
		// Get main photo URL for each bench
		if activity.BenchID != nil {
			photoURL := s.getMainPhotoURL(ctx, *activity.BenchID)
			if activityResponses[i].Bench != nil {
				activityResponses[i].Bench.MainPhotoURL = photoURL
			}
		}
	}

	return &responses.PaginatedActivitiesResponse{
		Activities: activityResponses,
		Pagination: responses.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// getMainPhotoURL fetches the main photo URL for a bench
func (s *activityService) getMainPhotoURL(ctx context.Context, benchID uint) *string {
	mainPhoto, err := s.photoRepo.GetMainPhoto(ctx, benchID)
	if err != nil || mainPhoto == nil {
		return nil
	}

	url := s.minioClient.GetPublicURL(mainPhoto.FilePathThumbnail)
	return &url
}
