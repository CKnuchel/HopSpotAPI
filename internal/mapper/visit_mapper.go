package mapper

import (
	"time"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
)

func CreateVisitRequestToDomain(req *requests.CreateVisitRequest, userID uint) *domain.Visit {
	visitedAt := time.Now()
	if req.VisitedAt != nil {
		visitedAt = *req.VisitedAt
	}

	return &domain.Visit{
		BenchID:   req.BenchID,
		UserID:    userID,
		VisitedAt: visitedAt,
		Comment:   req.Comment,
	}
}

func VisitToResponse(visit *domain.Visit) responses.VisitResponse {
	return responses.VisitResponse{
		ID: visit.ID,
		Bench: responses.VisitBenchResponse{
			ID:   visit.Bench.ID,
			Name: visit.Bench.Name,
		},
		VisitedAt: visit.VisitedAt,
		Comment:   visit.Comment,
		CreatedAt: visit.CreatedAt,
	}
}

func VisitsToListResponse(visits []domain.Visit) []responses.VisitResponse {
	result := make([]responses.VisitResponse, len(visits))
	for i, visit := range visits {
		result[i] = VisitToResponse(&visit)
	}
	return result
}
