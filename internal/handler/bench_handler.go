package handler

import (
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BenchHandler struct {
	benchService service.BenchService
}

func NewBenchHandler(benchService service.BenchService) *BenchHandler {
	return &BenchHandler{benchService: benchService}
}

// GET /api/v1/benches
func (h *BenchHandler) List(c *gin.Context) {
	var req requests.ListBenchesRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Define default pagination values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	// Call the service to get benches
	benches, err := h.benchService.List(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve benches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": benches})
}

// GET /api/v1/benches/:id
func (h *BenchHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bench ID"})
		return
	}

	// Call the service to get the bench by ID
	bench, err := h.benchService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bench"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": bench})
}

// POST /api/v1/benches
func (h *BenchHandler) Create(c *gin.Context) {
	// JWT Claims
	userId := c.MustGet(middleware.ContextKeyUserID).(uint)

	// Request data
	var req requests.CreateBenchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.benchService.Create(c.Request.Context(), &req, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bench"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// PATCH /api/v1/benches/:id
func (h *BenchHandler) Update(c *gin.Context) {
	// JWT Claims
	userId := c.MustGet(middleware.ContextKeyUserID).(uint)
	userRole := c.MustGet(middleware.ContextKeyUserRole).(domain.Role)
	isAdmin := userRole == domain.RoleAdmin

	// Request data
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bench ID"})
		return
	}

	var req requests.UpdateBenchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.benchService.Update(c.Request.Context(), uint(id), &req, userId, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bench"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// DELETE /api/v1/benches/:id
func (h *BenchHandler) Delete(c *gin.Context) {
	// JWT Claims
	userId := c.MustGet(middleware.ContextKeyUserID).(uint)
	userRole := c.MustGet(middleware.ContextKeyUserRole).(domain.Role)
	isAdmin := userRole == domain.RoleAdmin

	// Request data
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bench ID"})
		return
	}

	err = h.benchService.Delete(c.Request.Context(), uint(id), userId, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bench"})
		return
	}

	c.JSON(http.StatusOK, nil)
}
