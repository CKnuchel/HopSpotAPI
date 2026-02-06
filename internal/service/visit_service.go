package service

import (
	"context"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/apperror"
	"hopSpotAPI/pkg/logger"
	"hopSpotAPI/pkg/storage"
)

type VisitService interface {
	Create(ctx context.Context, req *requests.CreateVisitRequest, userID uint) (*responses.VisitResponse, error)
	List(ctx context.Context, req *requests.ListVisitsRequest, userID uint) (*responses.PaginatedVisitsResponse, error)
	GetCountBySpotID(ctx context.Context, spotID uint) (int64, error)
	Delete(ctx context.Context, visitID uint, userID uint) error
}

type visitService struct {
	visitRepo       repository.VisitRepository
	photoRepo       repository.PhotoRepository
	minioClient     *storage.MinioClient
	activityService ActivityService
}

func NewVisitService(visitRepo repository.VisitRepository, photoRepo repository.PhotoRepository, minioClient *storage.MinioClient, activityService ActivityService) VisitService {
	return &visitService{
		visitRepo:       visitRepo,
		photoRepo:       photoRepo,
		minioClient:     minioClient,
		activityService: activityService,
	}
}

func (v *visitService) GetCountBySpotID(ctx context.Context, spotID uint) (int64, error) {
	count, err := v.visitRepo.CountBySpotID(ctx, spotID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// List implements VisitService.
func (v *visitService) List(ctx context.Context, req *requests.ListVisitsRequest, userID uint) (*responses.PaginatedVisitsResponse, error) {
	filter := repository.VisitFilter{
		Page:      req.Page,
		Limit:     req.Limit,
		SortOrder: req.SortOrder,
		SpotID:    req.SpotID,
	}

	visits, total, err := v.visitRepo.FindByUserID(ctx, userID, filter)
	if err != nil {
		return nil, err
	}

	// calculate pagination info
	totalPages := int(total) / req.Limit
	if int(total)%req.Limit > 0 {
		totalPages++
	}

	// Map visits to responses with photo URLs
	visitResponses := make([]responses.VisitResponse, len(visits))
	for i, visit := range visits {
		visitResponses[i] = mapper.VisitToResponse(&visit)
		// Get main photo URL for each spot
		photoURL := v.getMainPhotoURL(ctx, visit.SpotID)
		visitResponses[i].Spot.MainPhotoURL = photoURL
	}

	return &responses.PaginatedVisitsResponse{
		Visits: visitResponses,
		Pagination: responses.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// Create implements VisitService.
func (v *visitService) Create(ctx context.Context, req *requests.CreateVisitRequest, userID uint) (*responses.VisitResponse, error) {
	visit := mapper.CreateVisitRequestToDomain(req, userID)

	if err := v.visitRepo.Create(ctx, visit); err != nil {
		return nil, err
	}

	// Reload visit with spot data
	visit, err := v.visitRepo.FindByID(ctx, visit.ID)
	if err != nil {
		return nil, err
	}

	response := mapper.VisitToResponse(visit)
	// Get main photo URL for the spot
	response.Spot.MainPhotoURL = v.getMainPhotoURL(ctx, visit.SpotID)

	// Create activity for visit (async)
	go func() {
		spotID := visit.SpotID
		if err := v.activityService.Create(context.Background(), userID, domain.ActionVisitAdded, &spotID); err != nil {
			logger.Warn().Err(err).Uint("spotID", visit.SpotID).Msg("failed to create visit_added activity")
		}
	}()

	return &response, nil
}

// getMainPhotoURL fetches the main photo URL for a spot
func (v *visitService) getMainPhotoURL(ctx context.Context, spotID uint) *string {
	mainPhoto, err := v.photoRepo.GetMainPhoto(ctx, spotID)
	if err != nil || mainPhoto == nil {
		return nil
	}

	url := v.minioClient.GetPublicURL(mainPhoto.FilePathThumbnail)
	return &url
}

// Delete deletes a visit if it belongs to the user
func (v *visitService) Delete(ctx context.Context, visitID uint, userID uint) error {
	// Check if visit exists and belongs to the user
	visit, err := v.visitRepo.FindByID(ctx, visitID)
	if err != nil {
		return err
	}

	if visit.UserID != userID {
		return apperror.ErrForbidden
	}

	return v.visitRepo.Delete(ctx, visitID)
}
