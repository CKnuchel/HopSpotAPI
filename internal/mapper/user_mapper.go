package mapper

import (
	"strconv"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
)

func RegisterRequestToUser(req *requests.RegisterRequest) *domain.User {
	return &domain.User{
		Email:       req.Email,
		DisplayName: req.DisplayName,
		// Password is hashed and set in the service layer
		// Role is set to default in the service layer
	}
}

func UserToResponse(user *domain.User) responses.UserResponse {
	return responses.UserResponse{
		ID:          strconv.Itoa(int(user.ID)),
		Email:       user.Email,
		DisplayName: user.DisplayName,
		Role:        string(user.Role),
		CreatedAt:   user.CreatedAt,
	}
}

func UsersToResponses(users []domain.User) []responses.UserResponse {
	responsesList := make([]responses.UserResponse, len(users))
	for i, user := range users {
		responsesList[i] = UserToResponse(&user)
	}
	return responsesList
}

func UsersToResponse(users []domain.User) []responses.UserResponse {
	responsesList := make([]responses.UserResponse, len(users))
	for i, user := range users {
		responsesList[i] = UserToResponse(&user)
	}
	return responsesList
}
