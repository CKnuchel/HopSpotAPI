package responses

import "time"

type VisitResponse struct {
	ID        uint              `json:"id"`
	Spot      VisitSpotResponse `json:"spot"`
	VisitedAt time.Time         `json:"visited_at"`
	Comment   string            `json:"comment,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

type VisitSpotResponse struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	MainPhotoURL *string `json:"main_photo_url,omitempty"`
}

type PaginatedVisitsResponse struct {
	Visits     []VisitResponse    `json:"visits"`
	Pagination PaginationResponse `json:"pagination"`
}

type VisitCountResponse struct {
	SpotID uint  `json:"spot_id"`
	Count  int64 `json:"count"`
}
