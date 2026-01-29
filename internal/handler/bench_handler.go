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
// ListBenches godoc
//
//	@Summary		List benches
//	@Description	Get a paginated list of benches with optional filters
//	@Tags			Benches
//	@Accept			json
//	@Produce		json
//	@Param			page			query		int		false	"Page number"					default(1)
//	@Param			limit			query		int		false	"Number of benches per page"	default(50)
//
// @ Param sort_by query string false "Sort by field" default("created_at") Enums(created_at, rating, distance)
//
//	@Param			sort_order		query		string	false	"Sort order"					Enums(asc, desc)	default("desc")
//	@Param			search			query		string	false	"Search term for bench name or location"
//	@Param			has_toilet		query		bool	false	"Filter by presence of toilet"
//	@Param			has_trash_bin	query		bool	false	"Filter by presence of trash bin"
//	@Param			min_rating		query		int		false	"Filter by minimum rating"
//	@Param			lat				query		number	false	"Latitude for proximity search"
//	@Param			lon				query		number	false	"Longitude for proximity search"
//	@Param			radius			query		int		false	"Radius in meters for proximity search"
//	@Success		200				{object}	responses.PaginatedBenchesResponse
//	@Failure		400
//	@Router			/api/v1/benches [get]
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
// GetBenchByID godoc
//
//	@Summary		Get bench by ID
//	@Description	Retrieve a single bench by its ID
//	@Tags			Benches
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Bench ID"
//	@Success		200	{object}	responses.BenchResponse
//	@Failure		400
//	@Router			/api/v1/benches/{id} [get]
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
// CreateBench godoc
//
//	@Summary		Create a new bench
//	@Description	Create a new bench with the provided details
//	@Tags			Benches
//	@Accept			json
//	@Produce		json
//	@Param			bench	body		requests.CreateBenchRequest	true	"Bench payload"
//	@Success		201		{object}	responses.BenchResponse
//	@Failure		400
//	@Router			/api/v1/benches [post]
func (h *BenchHandler) Create(c *gin.Context) {
	// JWT Claims
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	// Request data
	var req requests.CreateBenchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.benchService.Create(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bench"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// PATCH /api/v1/benches/:id
// UpdateBench godoc
//
//	@Summary		Update a bench
//	@Description	Update bench details by ID
//	@Tags			Benches
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"Bench ID"
//	@Param			bench	body		requests.UpdateBenchRequest	true	"Bench update payload"
//	@Success		200		{object}	responses.BenchResponse
//	@Failure		400
//	@Router			/api/v1/benches/{id} [patch]
func (h *BenchHandler) Update(c *gin.Context) {
	// JWT Claims
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}
	userRole, ok := c.MustGet(middleware.ContextKeyUserRole).(domain.Role)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid role context"})
		return
	}
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

	result, err := h.benchService.Update(c.Request.Context(), uint(id), &req, userID, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bench"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// DELETE /api/v1/benches/:id
// DeleteBench godoc
//
//	@Summary		Delete a bench
//	@Description	Delete a bench by ID
//	@Tags			Benches
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Bench ID"
//	@Success		200	{object}	nil
//	@Failure		400
//	@Router			/api/v1/benches/{id} [delete]
func (h *BenchHandler) Delete(c *gin.Context) {
	// JWT Claims
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}
	userRole, ok := c.MustGet(middleware.ContextKeyUserRole).(domain.Role)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid role context"})
		return
	}
	isAdmin := userRole == domain.RoleAdmin

	// Request data
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bench ID"})
		return
	}

	err = h.benchService.Delete(c.Request.Context(), uint(id), userID, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bench"})
		return
	}

	c.JSON(http.StatusOK, nil)
}
