package handler

import (
	"net/http"
	"strconv"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/service"
	"hopSpotAPI/pkg/apperror"

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
//	@Param			sort_by			query		string	false	"Sort by field"					Enums(name, rating, created_at, distance)	default(created_at)
//	@Param			sort_order		query		string	false	"Sort order"					Enums(asc, desc)							default(desc)
//	@Param			search			query		string	false	"Search term for bench name or description"
//	@Param			has_toilet		query		bool	false	"Filter by presence of toilet"
//	@Param			has_trash_bin	query		bool	false	"Filter by presence of trash bin"
//	@Param			min_rating		query		int		false	"Filter by minimum rating (1-5)"
//	@Param			lat				query		number	false	"Latitude for proximity search"
//	@Param			lon				query		number	false	"Longitude for proximity search"
//	@Param			radius			query		int		false	"Radius in meters for proximity search"
//	@Success		200				{object}	responses.PaginatedBenchesResponse
//	@Failure		400				{object}	apperror.ErrorResponse	"Bad Request"
//	@Failure		500				{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/benches [get]
func (h *BenchHandler) List(c *gin.Context) {
	var req requests.ListBenchesRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidRequest)
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
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": benches})
}

// GET /api/v1/benches/random
// GetRandomBench godoc
//
//	@Summary		Get a random bench
//	@Description	Retrieve a random bench from the database
//	@Tags			Benches
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.BenchResponse
//	@Failure		404	{object}	apperror.ErrorResponse	"No benches found"
//	@Failure		500	{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/benches/random [get]
func (h *BenchHandler) GetRandom(c *gin.Context) {
	bench, err := h.benchService.GetRandom(c.Request.Context())
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": bench})
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
//	@Failure		400	{object}	apperror.ErrorResponse	"Invalid bench ID"
//	@Failure		404	{object}	apperror.ErrorResponse	"Bench not found"
//	@Failure		500	{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/benches/{id} [get]
func (h *BenchHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	// Call the service to get the bench by ID
	bench, err := h.benchService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		apperror.RespondWithMappedError(c, err)
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
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			bench	body		requests.CreateBenchRequest	true	"Bench payload"
//	@Success		201		{object}	responses.BenchResponse
//	@Failure		400		{object}	apperror.ErrorResponse	"Bad Request"
//	@Failure		401		{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		500		{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/benches [post]
func (h *BenchHandler) Create(c *gin.Context) {
	// JWT Claims
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	// Request data
	var req requests.CreateBenchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidRequest)
		return
	}

	result, err := h.benchService.Create(c.Request.Context(), &req, userID)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// PATCH /api/v1/benches/:id
// UpdateBench godoc
//
//	@Summary		Update a bench
//	@Description	Update bench details by ID. Only the owner or an admin can update a bench.
//	@Tags			Benches
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"Bench ID"
//	@Param			bench	body		requests.UpdateBenchRequest	true	"Bench update payload"
//	@Success		200		{object}	responses.BenchResponse
//	@Failure		400		{object}	apperror.ErrorResponse	"Bad Request"
//	@Failure		401		{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		403		{object}	apperror.ErrorResponse	"Forbidden - not owner or admin"
//	@Failure		404		{object}	apperror.ErrorResponse	"Bench not found"
//	@Failure		500		{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/benches/{id} [patch]
func (h *BenchHandler) Update(c *gin.Context) {
	// JWT Claims
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}
	userRole, ok := c.MustGet(middleware.ContextKeyUserRole).(domain.Role)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}
	isAdmin := userRole == domain.RoleAdmin

	// Request data
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	var req requests.UpdateBenchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidRequest)
		return
	}

	result, err := h.benchService.Update(c.Request.Context(), uint(id), &req, userID, isAdmin)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// DELETE /api/v1/benches/:id
// DeleteBench godoc
//
//	@Summary		Delete a bench
//	@Description	Delete a bench by ID. Only the owner or an admin can delete a bench.
//	@Tags			Benches
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Bench ID"
//	@Success		200	{object}	nil	"Successfully deleted"
//	@Failure		400	{object}	apperror.ErrorResponse	"Invalid bench ID"
//	@Failure		401	{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	apperror.ErrorResponse	"Forbidden - not owner or admin"
//	@Failure		404	{object}	apperror.ErrorResponse	"Bench not found"
//	@Failure		500	{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/benches/{id} [delete]
func (h *BenchHandler) Delete(c *gin.Context) {
	// JWT Claims
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}
	userRole, ok := c.MustGet(middleware.ContextKeyUserRole).(domain.Role)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}
	isAdmin := userRole == domain.RoleAdmin

	// Request data
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	err = h.benchService.Delete(c.Request.Context(), uint(id), userID, isAdmin)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}
