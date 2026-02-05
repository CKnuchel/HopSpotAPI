package responses

import "time"

type VisitResponse struct {
	ID        uint               `json:"id"`
	Bench     VisitBenchResponse `json:"bench"`
	VisitedAt time.Time          `json:"visited_at"`
	Comment   string             `json:"comment,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
}

type VisitBenchResponse struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	MainPhotoURL *string `json:"main_photo_url,omitempty"`
}

type PaginatedVisitsResponse struct {
	Visits     []VisitResponse    `json:"visits"`
	Pagination PaginationResponse `json:"pagination"`
}

type VisitCountResponse struct {
	BenchID uint  `json:"bench_id"`
	Count   int64 `json:"count"`
}
