package responses

import "time"

type BenchResponse struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Latitude    float64      `json:"latitude"`
	Longitude   float64      `json:"longitude"`
	Description *string      `json:"description,omitempty"`
	Rating      *int         `json:"rating,omitempty"`
	HasToilet   bool         `json:"has_toilet"`
	HasTrashBin bool         `json:"has_trash_bin"`
	CreatedBy   UserResponse `json:"created_by"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type BenchListResponse struct {
	ID           uint     `json:"id"`
	Name         string   `json:"name"`
	Latitude     float64  `json:"latitude"`
	Longitude    float64  `json:"longitude"`
	Rating       *int     `json:"rating,omitempty"`
	HasToilet    bool     `json:"has_toilet"`
	HasTrashBin  bool     `json:"has_trash_bin"`
	MainPhotoURL *string  `json:"main_photo_url,omitempty"`
	Distance     *float64 `json:"distance,omitempty"` // Falls Koordinaten mitgegeben
}

type PaginatedBenchesResponse struct {
	Benches    []BenchListResponse `json:"benches"`
	Pagination PaginationResponse  `json:"pagination"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}
