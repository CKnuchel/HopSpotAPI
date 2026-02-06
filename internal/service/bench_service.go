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

type BenchService interface {
	Create(ctx context.Context, req *requests.CreateBenchRequest, userID uint) (*responses.BenchResponse, error)
	GetByID(ctx context.Context, id uint) (*responses.BenchResponse, error)
	GetRandom(ctx context.Context) (*responses.BenchResponse, error)
	List(ctx context.Context, req *requests.ListBenchesRequest) (*responses.PaginatedBenchesResponse, error)
	Update(ctx context.Context, id uint, req *requests.UpdateBenchRequest, userID uint, isAdmin bool) (*responses.BenchResponse, error)
	Delete(ctx context.Context, id uint, userID uint, isAdmin bool) error
}

type benchService struct {
	benchRepo           repository.BenchRepository
	photoRepo           repository.PhotoRepository
	minioClient         *storage.MinioClient
	notificationService NotificationService
	activityService     ActivityService
}

func NewBenchService(benchRepo repository.BenchRepository, photoRepo repository.PhotoRepository, minioClient *storage.MinioClient, notificationService NotificationService, activityService ActivityService) BenchService {
	return &benchService{
		benchRepo:           benchRepo,
		photoRepo:           photoRepo,
		minioClient:         minioClient,
		notificationService: notificationService,
		activityService:     activityService,
	}
}

// Create implements BenchService.
func (s *benchService) Create(ctx context.Context, req *requests.CreateBenchRequest, userID uint) (*responses.BenchResponse, error) {
	bench := mapper.CreateBenchRequestToDomain(req)
	bench.CreatedBy = userID

	if err := s.benchRepo.Create(ctx, bench); err != nil {
		return nil, err
	}

	// Reload mit Creator
	bench, err := s.benchRepo.FindByID(ctx, bench.ID)
	if err != nil {
		return nil, err
	}

	// Notify about new Bench (async)
	go func() {
		if err := s.notificationService.NotifyNewBench(ctx, bench, userID); err != nil {
			logger.Warn().Err(err).Uint("benchID", bench.ID).Msg("failed to send new bench notification")
		}
	}()

	// Create activity for bench creation (async)
	go func() {
		benchID := bench.ID
		if err := s.activityService.Create(context.Background(), userID, domain.ActionBenchCreated, &benchID); err != nil {
			logger.Warn().Err(err).Uint("benchID", bench.ID).Msg("failed to create bench_created activity")
		}
	}()

	response := mapper.BenchToResponse(bench)
	return &response, nil
}

// GetByID implements BenchService.
func (b *benchService) GetByID(ctx context.Context, id uint) (*responses.BenchResponse, error) {
	bench, err := b.benchRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if bench == nil {
		return nil, apperror.ErrBenchNotFound
	}

	response := mapper.BenchToResponse(bench)
	mainPhotoURL, err := b.getMainPhotoURL(ctx, id)
	if err != nil {
		logger.Warn().Err(err).Uint("benchID", id).Msg("failed to get main photo URL")
		// Continue without main photo
	}
	response.MainPhotoURL = mainPhotoURL

	return &response, nil
}

// GetRandom implements BenchService.
func (b *benchService) GetRandom(ctx context.Context) (*responses.BenchResponse, error) {
	bench, err := b.benchRepo.FindRandom(ctx)
	if err != nil {
		return nil, err
	}
	if bench == nil {
		return nil, apperror.ErrBenchNotFound
	}

	response := mapper.BenchToResponse(bench)
	mainPhotoURL, err := b.getMainPhotoURL(ctx, bench.ID)
	if err != nil {
		logger.Warn().Err(err).Uint("benchID", bench.ID).Msg("failed to get main photo URL")
		// Continue without main photo
	}
	response.MainPhotoURL = mainPhotoURL

	return &response, nil
}

func (s *benchService) List(ctx context.Context, req *requests.ListBenchesRequest) (*responses.PaginatedBenchesResponse, error) {
	// Define Filter
	filter := repository.BenchFilter{
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

	// Load Benches from Repo
	benches, _, err := s.benchRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// If coordinates are given: calculate distance + filter by radius
	benchResponses := make([]responses.BenchListResponse, 0, len(benches))

	for _, bench := range benches {
		resp := mapper.BenchToListResponse(&bench)

		mainPhotoURL, err := s.getMainPhotoURL(ctx, bench.ID)
		if err != nil {
			logger.Warn().Err(err).Uint("benchID", bench.ID).Msg("failed to get main photo URL")
		}

		resp.MainPhotoURL = mainPhotoURL

		if req.Lat != nil && req.Lon != nil {
			distance := utils.DistanceMeters(*req.Lat, *req.Lon, bench.Latitude, bench.Longitude)

			// Filter by Radius
			if req.Radius != nil && distance > float64(*req.Radius) {
				continue // Skip Benches outside the radius
			}

			resp.Distance = &distance
		}

		benchResponses = append(benchResponses, resp)
	}

	// Order by Distance if requested
	if req.SortBy == "distance" && req.Lat != nil && req.Lon != nil {
		sort.Slice(benchResponses, func(i, j int) bool {
			if req.SortOrder == "desc" {
				return *benchResponses[i].Distance > *benchResponses[j].Distance
			}
			return *benchResponses[i].Distance < *benchResponses[j].Distance
		})
	}

	// Pagination
	filteredTotal := int64(len(benchResponses))

	return &responses.PaginatedBenchesResponse{
		Benches: benchResponses,
		Pagination: responses.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      filteredTotal,
			TotalPages: int(math.Ceil(float64(filteredTotal) / float64(req.Limit))),
		},
	}, nil
}

// Update implements BenchService.
func (b *benchService) Update(ctx context.Context, id uint, req *requests.UpdateBenchRequest, userID uint, isAdmin bool) (*responses.BenchResponse, error) {
	// Get existing bench
	bench, err := b.benchRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if bench == nil {
		return nil, apperror.ErrBenchNotFound
	}

	// Check permissions
	if bench.CreatedBy != userID && !isAdmin {
		return nil, apperror.ErrForbidden
	}

	// Prepare fields to update
	if req.Name != nil {
		bench.Name = *req.Name
	}
	if req.Latitude != nil {
		bench.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		bench.Longitude = *req.Longitude
	}
	if req.Description != nil {
		bench.Description = *req.Description
	}
	if req.Rating != nil {
		bench.Rating = req.Rating
	}
	if req.HasToilet != nil {
		bench.HasToilet = *req.HasToilet
	}
	if req.HasTrashBin != nil {
		bench.HasTrashBin = *req.HasTrashBin
	}

	// Update bench
	if err := b.benchRepo.Update(ctx, bench); err != nil {
		return nil, err
	}

	response := mapper.BenchToResponse(bench)

	// Get main photo URL
	mainPhotoURL, err := b.getMainPhotoURL(ctx, id)
	if err != nil {
		logger.Warn().Err(err).Uint("benchID", id).Msg("failed to get main photo URL")
		// Continue without main photo
	}
	response.MainPhotoURL = mainPhotoURL

	return &response, nil
}

// Delete implements BenchService.
func (b *benchService) Delete(ctx context.Context, id uint, userID uint, isAdmin bool) error {
	// Get existing bench
	bench, err := b.benchRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if bench == nil {
		return apperror.ErrBenchNotFound
	}

	// Check permissions
	if bench.CreatedBy != userID && !isAdmin {
		return apperror.ErrForbidden
	}

	// Get all photos for this bench
	photos, err := b.photoRepo.FindByBenchID(ctx, id)
	if err != nil {
		logger.Warn().Err(err).Uint("benchID", id).Msg("failed to get photos for deletion")
		// Continue anyway - try to delete what we can
	}

	// Delete photos from MinIO storage
	for _, photo := range photos {
		if photo.FilePathOriginal != "" {
			if err := b.minioClient.Delete(ctx, photo.FilePathOriginal); err != nil {
				logger.Warn().Err(err).Str("path", photo.FilePathOriginal).Msg("failed to delete original from storage")
			}
		}
		if photo.FilePathMedium != "" {
			if err := b.minioClient.Delete(ctx, photo.FilePathMedium); err != nil {
				logger.Warn().Err(err).Str("path", photo.FilePathMedium).Msg("failed to delete medium from storage")
			}
		}
		if photo.FilePathThumbnail != "" {
			if err := b.minioClient.Delete(ctx, photo.FilePathThumbnail); err != nil {
				logger.Warn().Err(err).Str("path", photo.FilePathThumbnail).Msg("failed to delete thumbnail from storage")
			}
		}
	}

	// Delete photos from database
	for _, photo := range photos {
		if err := b.photoRepo.Delete(ctx, photo.ID); err != nil {
			logger.Warn().Err(err).Uint("photoID", photo.ID).Msg("failed to delete photo record")
		}
	}

	// Delete bench
	return b.benchRepo.Delete(ctx, id)
}

// getMainPhotoURL lädt das Hauptfoto einer Bank und generiert eine presigned URL
func (s *benchService) getMainPhotoURL(ctx context.Context, benchID uint) (*string, error) {
	mainPhoto, err := s.photoRepo.GetMainPhoto(ctx, benchID)
	if err != nil {
		return nil, err
	}
	if mainPhoto == nil {
		return nil, nil // No main photo
	}

	url := s.minioClient.GetPublicURL(mainPhoto.FilePathThumbnail)

	return &url, nil
}
