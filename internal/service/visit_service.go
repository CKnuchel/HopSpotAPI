package service

import (
	"context"

	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
	"hopSpotAPI/internal/repository"
)

type VisitService interface {
	Create(ctx context.Context, req *requests.CreateVisitRequest, userID uint) (*responses.VisitResponse, error)
	List(ctx context.Context, req *requests.ListVisitsRequest, userID uint) (*responses.PaginatedVisitsResponse, error)
	GetCountByBenchID(ctx context.Context, benchID uint) (int64, error)
}

type visitService struct {
	visitRepo repository.VisitRepository
}

func NewVisitService(visitRepo repository.VisitRepository) VisitService {
	return &visitService{visitRepo: visitRepo}
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

	return &responses.PaginatedVisitsResponse{
		Visits: mapper.VisitsToListResponse(visits),
		Pagination: responses.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}, err
}

// Create implements VisitService.
func (v *visitService) Create(ctx context.Context, req *requests.CreateVisitRequest, userID uint) (*responses.VisitResponse, error) {
	visit := mapper.CreateVisitRequestToDomain(req, userID)

	if err := v.visitRepo.Create(ctx, visit); err != nil {
		return nil, err
	}

	response := mapper.VisitToResponse(visit)
	return &response, nil
}
