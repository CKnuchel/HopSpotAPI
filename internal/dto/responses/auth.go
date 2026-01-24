package responses

import (
	"time"
)

type UserResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
}

type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}
