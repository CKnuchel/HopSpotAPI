package requests

type ListActivitiesRequest struct {
	Page       int     `form:"page,default=1" binding:"min=1"`
	Limit      int     `form:"limit,default=50" binding:"min=1,max=100"`
	ActionType *string `form:"action_type" binding:"omitempty,oneof=bench_created visit_added favorite_added"`
}
