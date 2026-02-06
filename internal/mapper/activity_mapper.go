package mapper

import (
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/responses"
)

func ActivityToResponse(activity *domain.Activity) responses.ActivityResponse {
	response := responses.ActivityResponse{
		ID:          activity.ID,
		ActionType:  activity.ActionType,
		Description: generateActivityDescription(activity),
		CreatedAt:   activity.CreatedAt,
		User: responses.ActivityUserResponse{
			ID:          activity.User.ID,
			DisplayName: activity.User.DisplayName,
		},
	}

	if activity.Spot != nil {
		response.Spot = &responses.ActivitySpotResponse{
			ID:   activity.Spot.ID,
			Name: activity.Spot.Name,
		}
	}

	return response
}

func ActivitiesToListResponse(activities []domain.Activity) []responses.ActivityResponse {
	result := make([]responses.ActivityResponse, len(activities))
	for i, activity := range activities {
		result[i] = ActivityToResponse(&activity)
	}
	return result
}

func generateActivityDescription(activity *domain.Activity) string {
	spotName := ""
	if activity.Spot != nil {
		spotName = activity.Spot.Name
	}

	switch activity.ActionType {
	case domain.ActionSpotCreated:
		return "hat " + spotName + " hinzugefügt"
	case domain.ActionVisitAdded:
		return "hat " + spotName + " besucht"
	case domain.ActionFavoriteAdded:
		return "hat " + spotName + " als Favorit markiert"
	default:
		return ""
	}
}
