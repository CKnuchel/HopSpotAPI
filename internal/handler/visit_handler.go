package handler

import (
	"net/http"
	"strconv"

	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/service"
	"hopSpotAPI/pkg/apperror"

	"github.com/gin-gonic/gin"
)

type VisitHandler struct {
	visitService service.VisitService
}

func NewVisitHandler(visitService service.VisitService) *VisitHandler {
	return &VisitHandler{visitService: visitService}
}

// GET /api/v1/spots/:id/visits/count
// GetVisitCountBySpotID godoc
//
//	@Summary		Get visit count by spot ID
//	@Description	Retrieve the total number of visits for a specific spot
//	@Tags			Visits
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Spot ID"
//	@Success		200	{object}	responses.VisitCountResponse
//	@Failure		400	{object}	apperror.ErrorResponse
//	@Router			/api/v1/spots/{id}/visits/count [get]
func (h *VisitHandler) GetVisitCountBySpotID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	count, err := h.visitService.GetCountBySpotID(c.Request.Context(), uint(id))
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, responses.VisitCountResponse{SpotID: uint(id), Count: count})
}

// GET /api/v1/visits
// ListVisits godoc
//
//	@Summary		List visits
//	@Description	Get a paginated list of visits for the authenticated user
//	@Tags			Visits
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int	false	"Page number"
//	@Param			limit	query		int	false	"Number of items per page"
//	@Success		200		{object}	responses.PaginatedVisitsResponse
//	@Failure		400		{object}	apperror.ErrorResponse
//	@Router			/api/v1/visits [get]
func (h *VisitHandler) ListVisits(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	var req = requests.ListVisitsRequest{}
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

	responses, err := h.visitService.List(c.Request.Context(), &req, userID)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, responses)
}

// POST /api/v1/visits
// CreateVisit godoc
//
//	@Summary		Create a new visit
//	@Description	Record a new visit to a spot
//	@Tags			Visits
//	@Accept			json
//	@Produce		json
//	@Param			visit	body		requests.CreateVisitRequest	true	"Visit payload"
//	@Success		201		{object}	responses.VisitResponse
//	@Failure		400		{object}	apperror.ErrorResponse
//	@Router			/api/v1/visits [post]
func (h *VisitHandler) CreateVisit(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	var req requests.CreateVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidRequest)
		return
	}

	response, err := h.visitService.Create(c.Request.Context(), &req, userID)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// DELETE /api/v1/visits/:id
// DeleteVisit godoc
//
//	@Summary		Delete a visit
//	@Description	Delete a visit by ID (only own visits can be deleted)
//	@Tags			Visits
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Visit ID"
//	@Success		204
//	@Failure		400	{object}	apperror.ErrorResponse
//	@Failure		403	{object}	apperror.ErrorResponse
//	@Failure		404	{object}	apperror.ErrorResponse
//	@Router			/api/v1/visits/{id} [delete]
func (h *VisitHandler) DeleteVisit(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	if err := h.visitService.Delete(c.Request.Context(), uint(id), userID); err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
