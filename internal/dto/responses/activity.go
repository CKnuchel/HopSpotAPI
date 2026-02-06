package responses

import "time"

type ActivityResponse struct {
	ID          uint                   `json:"id"`
	ActionType  string                 `json:"action_type"`
	User        ActivityUserResponse   `json:"user"`
	Bench       *ActivityBenchResponse `json:"bench,omitempty"`
	Description string                 `json:"description"`
	CreatedAt   time.Time              `json:"created_at"`
}

type ActivityUserResponse struct {
	ID          uint   `json:"id"`
	DisplayName string `json:"display_name"`
}

type ActivityBenchResponse struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	MainPhotoURL *string `json:"main_photo_url,omitempty"`
}

type PaginatedActivitiesResponse struct {
	Activities []ActivityResponse `json:"activities"`
	Pagination PaginationResponse `json:"pagination"`
}
