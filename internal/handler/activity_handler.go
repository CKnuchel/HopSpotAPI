package handler

import (
	"net/http"

	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/service"
	"hopSpotAPI/pkg/apperror"

	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	activityService service.ActivityService
}

func NewActivityHandler(activityService service.ActivityService) *ActivityHandler {
	return &ActivityHandler{activityService: activityService}
}

// GET /api/v1/activities
// List godoc
//
//	@Summary		List activities
//	@Description	Get a paginated list of activities (activity feed)
//	@Tags			Activities
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"
//	@Param			limit		query		int		false	"Number of items per page"
//	@Param			action_type	query		string	false	"Filter by action type (bench_created, visit_added, favorite_added)"
//	@Success		200			{object}	responses.PaginatedActivitiesResponse
//	@Failure		400			{object}	apperror.ErrorResponse
//	@Router			/api/v1/activities [get]
func (h *ActivityHandler) List(c *gin.Context) {
	var req = requests.ListActivitiesRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidRequest)
		return
	}

	// Defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}

	response, err := h.activityService.List(c.Request.Context(), &req)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
