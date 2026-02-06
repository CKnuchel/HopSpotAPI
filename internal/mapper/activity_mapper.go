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

	if activity.Bench != nil {
		response.Bench = &responses.ActivityBenchResponse{
			ID:   activity.Bench.ID,
			Name: activity.Bench.Name,
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
	benchName := ""
	if activity.Bench != nil {
		benchName = activity.Bench.Name
	}

	switch activity.ActionType {
	case domain.ActionBenchCreated:
		return "hat " + benchName + " hinzugefügt"
	case domain.ActionVisitAdded:
		return "hat " + benchName + " besucht"
	case domain.ActionFavoriteAdded:
		return "hat " + benchName + " als Favorit markiert"
	default:
		return ""
	}
}
