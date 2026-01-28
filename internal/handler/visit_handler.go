package handler

import (
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/dto/responses"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type VisitHandler struct {
	visitService service.VisitService
}

func NewVisitHandler(visitService service.VisitService) *VisitHandler {
	return &VisitHandler{visitService: visitService}
}

// GET /api/v1/benches/:id/visits/count
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
func (h *VisitHandler) ListVisits(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uint)

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
func (h *VisitHandler) CreateVisit(c *gin.Context) {
	userID := c.MustGet(middleware.ContextKeyUserID).(uint)

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
