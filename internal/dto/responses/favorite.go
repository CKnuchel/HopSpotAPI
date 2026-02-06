package responses

import "time"

type FavoriteResponse struct {
	ID        uint                 `json:"id"`
	Spot      FavoriteSpotResponse `json:"spot"`
	CreatedAt time.Time            `json:"created_at"`
}

type FavoriteSpotResponse struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Rating       *int    `json:"rating,omitempty"`
	HasToilet    bool    `json:"has_toilet"`
	HasTrashBin  bool    `json:"has_trash_bin"`
	MainPhotoURL *string `json:"main_photo_url,omitempty"`
}

type PaginatedFavoritesResponse struct {
	Favorites  []FavoriteResponse `json:"favorites"`
	Pagination PaginationResponse `json:"pagination"`
}
