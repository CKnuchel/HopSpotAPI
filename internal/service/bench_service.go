package service

import (
	"context"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/mapper"
	"hopSpotAPI/internal/repository"
	"hopSpotAPI/pkg/apperror"
	"hopSpotAPI/pkg/utils"
	"math"
	"sort"
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

	// Load Benches from Repo
	benches, _, err := s.benchRepo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// If coordinates are given: calculate distance + filter by radius
	var benchResponses []responses.BenchListResponse

	for _, bench := range benches {
		resp := mapper.BenchToListResponse(&bench)

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
