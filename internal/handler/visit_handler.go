package handler

import (
	"net/http"
	"strconv"

	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/service"

	"github.com/gin-gonic/gin"
)

type VisitHandler struct {
	visitService service.VisitService
}

func NewVisitHandler(visitService service.VisitService) *VisitHandler {
	return &VisitHandler{visitService: visitService}
}

// GET /api/v1/benches/:id/visits/count
// GetVisitCountByBenchID godoc
//
//	@Summary		Get visit count by bench ID
//	@Description	Retrieve the total number of visits for a specific bench
//	@Tags			Visits
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Bench ID"
//	@Success		200	{object}	responses.VisitCountResponse
//	@Failure		400
//	@Router			/api/v1/benches/{id}/visits/count [get]
func (h *VisitHandler) GetVisitCountByBenchID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bench ID"})
		return
	}

	count, err := h.visitService.GetCountByBenchID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve visit count"})
		return
	}

	c.JSON(http.StatusOK, responses.VisitCountResponse{BenchID: uint(id), Count: count})
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
//	@Failure		400
//	@Router			/api/v1/visits [get]
func (h *VisitHandler) ListVisits(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	var req = requests.ListVisitsRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve visits"})
		return
	}

	c.JSON(http.StatusOK, responses)
}

// POST /api/v1/visits
// CreateVisit godoc
//
//	@Summary		Create a new visit
//	@Description	Record a new visit to a bench
//	@Tags			Visits
//	@Accept			json
//	@Produce		json
//	@Param			visit	body		requests.CreateVisitRequest	true	"Visit payload"
//	@Success		201		{object}	responses.VisitResponse
//	@Failure		400
//	@Router			/api/v1/visits [post]
func (h *VisitHandler) CreateVisit(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	var req requests.CreateVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.visitService.Create(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create visit"})
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
//	@Failure		400
//	@Failure		403
//	@Failure		404
//	@Router			/api/v1/visits/{id} [delete]
func (h *VisitHandler) DeleteVisit(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visit ID"})
		return
	}

	if err := h.visitService.Delete(c.Request.Context(), uint(id), userID); err != nil {
		if err.Error() == "forbidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own visits"})
			return
		}
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Visit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete visit"})
		return
	}

	c.Status(http.StatusNoContent)
}
