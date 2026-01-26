package mapper

import (
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
)

func CreateBenchRequestToDomain(req *requests.CreateBenchRequest) *domain.Bench {
	return &domain.Bench{
		Name:        req.Name,
		Description: *req.Description,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Rating:      req.Rating,
		HasToilet:   req.HasToilet,
		HasTrashBin: req.HasTrashBin,
	}
}

func BenchToResponse(bench *domain.Bench) responses.BenchResponse {
	return responses.BenchResponse{
		ID:          bench.ID,
		Name:        bench.Name,
		Latitude:    bench.Latitude,
		Longitude:   bench.Longitude,
		Description: &bench.Description,
		Rating:      bench.Rating,
		HasToilet:   bench.HasToilet,
		HasTrashBin: bench.HasTrashBin,
		CreatedBy:   UserToResponse(&bench.Creator),
		CreatedAt:   bench.CreatedAt,
		UpdatedAt:   bench.UpdatedAt,
	}
}

func BenchToListResponse(bench *domain.Bench) responses.BenchListResponse {
	return responses.BenchListResponse{
		ID:          bench.ID,
		Name:        bench.Name,
		Latitude:    bench.Latitude,
		Longitude:   bench.Longitude,
		Rating:      bench.Rating,
		HasToilet:   bench.HasToilet,
		HasTrashBin: bench.HasTrashBin,
		// MainPhotoURL und Distance werden separat gesetzt
	}
}

func BenchesToListResponse(benches []domain.Bench) []responses.BenchListResponse {
	result := make([]responses.BenchListResponse, len(benches))
	for i, bench := range benches {
		result[i] = BenchToListResponse(&bench)
	}
	return result
}
