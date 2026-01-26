package service

import (
	"context"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/apperror"
)

type BenchService interface {
	Create(ctx context.Context, req *requests.CreateBenchRequest, userID uint) (*responses.BenchResponse, error)
	GetByID(ctx context.Context, id uint) (*responses.BenchResponse, error)
	List(ctx context.Context, req *requests.ListBenchesRequest) (*responses.PaginatedBenchesResponse, error)
	Update(ctx context.Context, id uint, req *requests.UpdateBenchRequest, userID uint, isAdmin bool) (*responses.BenchResponse, error)
	Delete(ctx context.Context, id uint, userID uint, isAdmin bool) error
}

type benchService struct {
	benchRepo repository.BenchRepository
}

func NewBenchService(benchRepo repository.BenchRepository) BenchService {
	return &benchService{benchRepo: benchRepo}
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
	return &response, nil
}

// List implements BenchService.
func (b *benchService) List(ctx context.Context, req *requests.ListBenchesRequest) (*responses.PaginatedBenchesResponse, error) {
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

	// Find
	benches, total, err := b.benchRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Pagination berechnen
	totalPages := int(total) / req.Limit
	if int(total)%req.Limit > 0 {
		totalPages++
	}

	return &responses.PaginatedBenchesResponse{
		Benches: mapper.BenchesToListResponse(benches),
		Pagination: responses.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      total,
			TotalPages: totalPages,
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

	// Delete bench
	return b.benchRepo.Delete(ctx, id)
}
