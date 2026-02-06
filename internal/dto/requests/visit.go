package requests

import "time"

type CreateVisitRequest struct {
	SpotID    uint       `json:"spot_id" binding:"required"`
	VisitedAt *time.Time `json:"visited_at" default:"now"`
	Comment   string     `json:"comment" binding:"max=500"`
}

type ListVisitsRequest struct {
	Page      int    `form:"page,default=1" binding:"min=1"`
	Limit     int    `form:"limit,default=50" binding:"min=1,max=100"`
	SpotID    *uint  `form:"spot_id"`
	SortOrder string `form:"sort_order,default=desc" binding:"omitempty,oneof=asc desc"`
}
