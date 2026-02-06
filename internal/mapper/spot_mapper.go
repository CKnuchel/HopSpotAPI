package mapper

import (
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
)

func CreateSpotRequestToDomain(req *requests.CreateSpotRequest) *domain.Spot {
	spot := &domain.Spot{
		Name:        req.Name,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		HasToilet:   req.HasToilet,
		HasTrashBin: req.HasTrashBin,
	}

	// Handle optional fields with nil checks
	if req.Description != nil {
		spot.Description = *req.Description
	}

	if req.Rating != nil {
		spot.Rating = req.Rating
	}

	return spot
}

func SpotToResponse(spot *domain.Spot) responses.SpotResponse {
	return responses.SpotResponse{
		ID:          spot.ID,
		Name:        spot.Name,
		Latitude:    spot.Latitude,
		Longitude:   spot.Longitude,
		Description: &spot.Description,
		Rating:      spot.Rating,
		HasToilet:   spot.HasToilet,
		HasTrashBin: spot.HasTrashBin,
		CreatedBy:   UserToResponse(&spot.Creator),
		CreatedAt:   spot.CreatedAt,
		UpdatedAt:   spot.UpdatedAt,
	}
}

func SpotToListResponse(spot *domain.Spot) responses.SpotListResponse {
	return responses.SpotListResponse{
		ID:          spot.ID,
		Name:        spot.Name,
		Latitude:    spot.Latitude,
		Longitude:   spot.Longitude,
		Rating:      spot.Rating,
		HasToilet:   spot.HasToilet,
		HasTrashBin: spot.HasTrashBin,
		// MainPhotoURL und Distance werden separat gesetzt
	}
}

func SpotsToListResponse(spots []domain.Spot) []responses.SpotListResponse {
	result := make([]responses.SpotListResponse, len(spots))
	for i, spot := range spots {
		result[i] = SpotToListResponse(&spot)
	}
	return result
}
