package service

import (
	"context"

	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/storage"
)

type VisitService interface {
	Create(ctx context.Context, req *requests.CreateVisitRequest, userID uint) (*responses.VisitResponse, error)
	List(ctx context.Context, req *requests.ListVisitsRequest, userID uint) (*responses.PaginatedVisitsResponse, error)
	GetCountByBenchID(ctx context.Context, benchID uint) (int64, error)
}

type visitService struct {
	visitRepo   repository.VisitRepository
	photoRepo   repository.PhotoRepository
	minioClient *storage.MinioClient
}

func NewVisitService(visitRepo repository.VisitRepository, photoRepo repository.PhotoRepository, minioClient *storage.MinioClient) VisitService {
	return &visitService{
		visitRepo:   visitRepo,
		photoRepo:   photoRepo,
		minioClient: minioClient,
	}
}

func (v *visitService) GetCountByBenchID(ctx context.Context, benchID uint) (int64, error) {
	count, err := v.visitRepo.CountByBenchID(ctx, benchID)
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
		BenchID:   req.BenchID,
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
		// Get main photo URL for each bench
		photoURL := v.getMainPhotoURL(ctx, visit.BenchID)
		visitResponses[i].Bench.MainPhotoURL = photoURL
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

	// Reload visit with bench data
	visit, err := v.visitRepo.FindByID(ctx, visit.ID)
	if err != nil {
		return nil, err
	}

	response := mapper.VisitToResponse(visit)
	// Get main photo URL for the bench
	response.Bench.MainPhotoURL = v.getMainPhotoURL(ctx, visit.BenchID)
	return &response, nil
}

// getMainPhotoURL fetches the main photo URL for a bench
func (v *visitService) getMainPhotoURL(ctx context.Context, benchID uint) *string {
	mainPhoto, err := v.photoRepo.GetMainPhoto(ctx, benchID)
	if err != nil || mainPhoto == nil {
		return nil
	}

	url := v.minioClient.GetPublicURL(mainPhoto.FilePathThumbnail)
	return &url
}
