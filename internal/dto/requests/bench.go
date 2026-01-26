package requests

type CreateBenchRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Latitude    float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude   float64 `json:"longitude" binding:"required,min=-180,max=180"`
	Description *string `json:"description" binding:"omitempty,max=5000"`
	Rating      *int    `json:"rating" binding:"omitempty,min=1,max=5"`
	HasToilet   bool    `json:"has_toilet"`
	HasTrashBin bool    `json:"has_trash_bin"`
}

type UpdateBenchRequest struct {
	Name        *string `json:"name" binding:"omitempty,min=1,max=255"`
	Description *string `json:"description" binding:"omitempty,max=5000"`
	Rating      *int    `json:"rating" binding:"omitempty,min=1,max=5"`
	HasToilet   *bool   `json:"has_toilet"`
	HasTrashBin *bool   `json:"has_trash_bin"`
}

type ListBenchesRequest struct {
	Page        int      `form:"page,default=1" binding:"min=1"`
	Limit       int      `form:"limit,default=50" binding:"min=1,max=100"`
	SortBy      string   `form:"sort_by,default=created_at" binding:"omitempty,oneof=name rating created_at distance"`
	SortOrder   string   `form:"sort_order,default=desc" binding:"omitempty,oneof=asc desc"`
	HasToilet   *bool    `form:"has_toilet"`
	HasTrashBin *bool    `form:"has_trash_bin"`
	MinRating   *int     `form:"min_rating" binding:"omitempty,min=1,max=5"`
	Search      string   `form:"search"`
	Lat         *float64 `form:"lat"`
	Lon         *float64 `form:"lon"`
	Radius      *int     `form:"radius"` // in Metern
}
