package service

import (
	"context"
	"math"
	"sort"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/apperror"
	"hopSpotAPI/pkg/logger"
	"hopSpotAPI/pkg/storage"
	"hopSpotAPI/pkg/utils"
)

type SpotService interface {
	Create(ctx context.Context, req *requests.CreateSpotRequest, userID uint) (*responses.SpotResponse, error)
	GetByID(ctx context.Context, id uint) (*responses.SpotResponse, error)
	GetRandom(ctx context.Context) (*responses.SpotResponse, error)
	List(ctx context.Context, req *requests.ListSpotsRequest) (*responses.PaginatedSpotsResponse, error)
	Update(ctx context.Context, id uint, req *requests.UpdateSpotRequest, userID uint, isAdmin bool) (*responses.SpotResponse, error)
	Delete(ctx context.Context, id uint, userID uint, isAdmin bool) error
}

type spotService struct {
	spotRepo            repository.SpotRepository
	photoRepo           repository.PhotoRepository
	minioClient         *storage.MinioClient
	notificationService NotificationService
	activityService     ActivityService
}

func NewSpotService(spotRepo repository.SpotRepository, photoRepo repository.PhotoRepository, minioClient *storage.MinioClient, notificationService NotificationService, activityService ActivityService) SpotService {
	return &spotService{
		spotRepo:            spotRepo,
		photoRepo:           photoRepo,
		minioClient:         minioClient,
		notificationService: notificationService,
		activityService:     activityService,
	}
}

// Create implements SpotService.
func (s *spotService) Create(ctx context.Context, req *requests.CreateSpotRequest, userID uint) (*responses.SpotResponse, error) {
	spot := mapper.CreateSpotRequestToDomain(req)
	spot.CreatedBy = userID

	if err := s.spotRepo.Create(ctx, spot); err != nil {
		return nil, err
	}

	// Reload mit Creator
	spot, err := s.spotRepo.FindByID(ctx, spot.ID)
	if err != nil {
		return nil, err
	}

	// Notify about new Spot (async)
	go func() {
		if err := s.notificationService.NotifyNewSpot(context.Background(), spot, userID); err != nil {
			logger.Warn().Err(err).Uint("spotID", spot.ID).Msg("failed to send new spot notification")
		}
	}()

	// Create activity for spot creation (async)
	go func() {
		if s.activityService != nil {
			spotID := spot.ID
			if err := s.activityService.Create(context.Background(), userID, domain.ActionSpotCreated, &spotID); err != nil {
				logger.Warn().Err(err).Uint("spotID", spot.ID).Msg("failed to create spot_created activity")
			}
		}
	}()

	response := mapper.SpotToResponse(spot)
	return &response, nil
}

// GetByID implements SpotService.
func (s *spotService) GetByID(ctx context.Context, id uint) (*responses.SpotResponse, error) {
	spot, err := s.spotRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if spot == nil {
		return nil, apperror.ErrSpotNotFound
	}

	response := mapper.SpotToResponse(spot)
	mainPhotoURL, err := s.getMainPhotoURL(ctx, id)
	if err != nil {
		logger.Warn().Err(err).Uint("spotID", id).Msg("failed to get main photo URL")
		// Continue without main photo
	}
	response.MainPhotoURL = mainPhotoURL

	return &response, nil
}

// GetRandom implements SpotService.
func (s *spotService) GetRandom(ctx context.Context) (*responses.SpotResponse, error) {
	spot, err := s.spotRepo.FindRandom(ctx)
	if err != nil {
		return nil, err
	}
	if spot == nil {
		return nil, apperror.ErrSpotNotFound
	}

	response := mapper.SpotToResponse(spot)
	mainPhotoURL, err := s.getMainPhotoURL(ctx, spot.ID)
	if err != nil {
		logger.Warn().Err(err).Uint("spotID", spot.ID).Msg("failed to get main photo URL")
		// Continue without main photo
	}
	response.MainPhotoURL = mainPhotoURL

	return &response, nil
}

func (s *spotService) List(ctx context.Context, req *requests.ListSpotsRequest) (*responses.PaginatedSpotsResponse, error) {
	// Define Filter
	filter := repository.SpotFilter{
		Page:        req.Page,
		Limit:       req.Limit,
		SortBy:      req.SortBy,
		SortOrder:   req.SortOrder,
		HasToilet:   req.HasToilet,
		HasTrashBin: req.HasTrashBin,
		MinRating:   req.MinRating,
		Search:      req.Search,
		Lat:         req.Lat,
		Lon:         req.Lon,
		Radius:      req.Radius,
	}

	if req.Lat != nil && req.Lon != nil {
		if err := utils.ValidateCoordinates(*req.Lat, *req.Lon); err != nil {
			return nil, err
		}
	}

	// Load Spots from Repo
	spots, _, err := s.spotRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// If coordinates are given: calculate distance + filter by radius
	spotResponses := make([]responses.SpotListResponse, 0, len(spots))

	for _, spot := range spots {
		resp := mapper.SpotToListResponse(&spot)

		mainPhotoURL, err := s.getMainPhotoURL(ctx, spot.ID)
		if err != nil {
			logger.Warn().Err(err).Uint("spotID", spot.ID).Msg("failed to get main photo URL")
		}

		resp.MainPhotoURL = mainPhotoURL

		if req.Lat != nil && req.Lon != nil {
			distance := utils.DistanceMeters(*req.Lat, *req.Lon, spot.Latitude, spot.Longitude)

			// Filter by Radius
			if req.Radius != nil && distance > float64(*req.Radius) {
				continue // Skip Spots outside the radius
			}

			resp.Distance = &distance
		}

		spotResponses = append(spotResponses, resp)
	}

	// Order by Distance if requested
	if req.SortBy == "distance" && req.Lat != nil && req.Lon != nil {
		sort.Slice(spotResponses, func(i, j int) bool {
			if req.SortOrder == "desc" {
				return *spotResponses[i].Distance > *spotResponses[j].Distance
			}
			return *spotResponses[i].Distance < *spotResponses[j].Distance
		})
	}

	// Pagination
	filteredTotal := int64(len(spotResponses))

	return &responses.PaginatedSpotsResponse{
		Spots: spotResponses,
		Pagination: responses.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      filteredTotal,
			TotalPages: int(math.Ceil(float64(filteredTotal) / float64(req.Limit))),
		},
	}, nil
}

// Update implements SpotService.
func (s *spotService) Update(ctx context.Context, id uint, req *requests.UpdateSpotRequest, userID uint, isAdmin bool) (*responses.SpotResponse, error) {
	// Get existing spot
	spot, err := s.spotRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if spot == nil {
		return nil, apperror.ErrSpotNotFound
	}

	// Check permissions
	if spot.CreatedBy != userID && !isAdmin {
		return nil, apperror.ErrForbidden
	}

	// Prepare fields to update
	if req.Name != nil {
		spot.Name = *req.Name
	}
	if req.Latitude != nil {
		spot.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		spot.Longitude = *req.Longitude
	}
	if req.Description != nil {
		spot.Description = *req.Description
	}
	if req.Rating != nil {
		spot.Rating = req.Rating
	}
	if req.HasToilet != nil {
		spot.HasToilet = *req.HasToilet
	}
	if req.HasTrashBin != nil {
		spot.HasTrashBin = *req.HasTrashBin
	}

	// Update spot
	if err := s.spotRepo.Update(ctx, spot); err != nil {
		return nil, err
	}

	response := mapper.SpotToResponse(spot)

	// Get main photo URL
	mainPhotoURL, err := s.getMainPhotoURL(ctx, id)
	if err != nil {
		logger.Warn().Err(err).Uint("spotID", id).Msg("failed to get main photo URL")
		// Continue without main photo
	}
	response.MainPhotoURL = mainPhotoURL

	return &response, nil
}

// Delete implements SpotService.
func (s *spotService) Delete(ctx context.Context, id uint, userID uint, isAdmin bool) error {
	// Get existing spot
	spot, err := s.spotRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if spot == nil {
		return apperror.ErrSpotNotFound
	}

	// Check permissions
	if spot.CreatedBy != userID && !isAdmin {
		return apperror.ErrForbidden
	}

	// Get all photos for this spot (including soft-deleted ones)
	photos, err := s.photoRepo.FindBySpotIDUnscoped(ctx, id)
	if err != nil {
		logger.Warn().Err(err).Uint("spotID", id).Msg("failed to get photos for deletion")
		// Continue anyway - try to delete what we can
	}

	// Delete photos from MinIO storage
	for _, photo := range photos {
		if photo.FilePathOriginal != "" {
			if err := s.minioClient.Delete(ctx, photo.FilePathOriginal); err != nil {
				logger.Warn().Err(err).Str("path", photo.FilePathOriginal).Msg("failed to delete original from storage")
			}
		}
		if photo.FilePathMedium != "" {
			if err := s.minioClient.Delete(ctx, photo.FilePathMedium); err != nil {
				logger.Warn().Err(err).Str("path", photo.FilePathMedium).Msg("failed to delete medium from storage")
			}
		}
		if photo.FilePathThumbnail != "" {
			if err := s.minioClient.Delete(ctx, photo.FilePathThumbnail); err != nil {
				logger.Warn().Err(err).Str("path", photo.FilePathThumbnail).Msg("failed to delete thumbnail from storage")
			}
		}
	}

	// Delete photos from database
	for _, photo := range photos {
		if err := s.photoRepo.HardDelete(ctx, photo.ID); err != nil {
			logger.Warn().Err(err).Uint("photoID", photo.ID).Msg("failed to delete photo record")
		}
	}

	// Delete spot
	return s.spotRepo.Delete(ctx, id)
}

// getMainPhotoURL lädt das Hauptfoto eines Spots und generiert eine presigned URL
func (s *spotService) getMainPhotoURL(ctx context.Context, spotID uint) (*string, error) {
	mainPhoto, err := s.photoRepo.GetMainPhoto(ctx, spotID)
	if err != nil {
		return nil, err
	}
	if mainPhoto == nil {
		return nil, nil // No main photo
	}

	url := s.minioClient.GetPublicURL(mainPhoto.FilePathThumbnail)

	return &url, nil
}
