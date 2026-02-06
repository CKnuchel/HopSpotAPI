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

type SpotHandler struct {
	spotService service.SpotService
}

func NewSpotHandler(spotService service.SpotService) *SpotHandler {
	return &SpotHandler{spotService: spotService}
}

// GET /api/v1/spots
// ListSpots godoc
//
//	@Summary		List spots
//	@Description	Get a paginated list of spots with optional filters
//	@Tags			Spots
//	@Accept			json
//	@Produce		json
//	@Param			page			query		int		false	"Page number"					default(1)
//	@Param			limit			query		int		false	"Number of spots per page"	default(50)
//	@Param			sort_by			query		string	false	"Sort by field"					Enums(name, rating, created_at, distance)	default(created_at)
//	@Param			sort_order		query		string	false	"Sort order"					Enums(asc, desc)							default(desc)
//	@Param			search			query		string	false	"Search term for spot name or description"
//	@Param			has_toilet		query		bool	false	"Filter by presence of toilet"
//	@Param			has_trash_bin	query		bool	false	"Filter by presence of trash bin"
//	@Param			min_rating		query		int		false	"Filter by minimum rating (1-5)"
//	@Param			lat				query		number	false	"Latitude for proximity search"
//	@Param			lon				query		number	false	"Longitude for proximity search"
//	@Param			radius			query		int		false	"Radius in meters for proximity search"
//	@Success		200				{object}	responses.PaginatedSpotsResponse
//	@Failure		400				{object}	apperror.ErrorResponse	"Bad Request"
//	@Failure		500				{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/spots [get]
func (h *SpotHandler) List(c *gin.Context) {
	var req requests.ListSpotsRequest

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

	// Call the service to get spots
	spots, err := h.spotService.List(c.Request.Context(), &req)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": spots})
}

// GET /api/v1/spots/random
// GetRandomSpot godoc
//
//	@Summary		Get a random spot
//	@Description	Retrieve a random spot from the database
//	@Tags			Spots
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.SpotResponse
//	@Failure		404	{object}	apperror.ErrorResponse	"No spots found"
//	@Failure		500	{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/spots/random [get]
func (h *SpotHandler) GetRandom(c *gin.Context) {
	spot, err := h.spotService.GetRandom(c.Request.Context())
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": spot})
}

// GET /api/v1/spots/:id
// GetSpotByID godoc
//
//	@Summary		Get spot by ID
//	@Description	Retrieve a single spot by its ID
//	@Tags			Spots
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Spot ID"
//	@Success		200	{object}	responses.SpotResponse
//	@Failure		400	{object}	apperror.ErrorResponse	"Invalid spot ID"
//	@Failure		404	{object}	apperror.ErrorResponse	"Spot not found"
//	@Failure		500	{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/spots/{id} [get]
func (h *SpotHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	// Call the service to get the spot by ID
	spot, err := h.spotService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": spot})
}

// POST /api/v1/spots
// CreateSpot godoc
//
//	@Summary		Create a new spot
//	@Description	Create a new spot with the provided details
//	@Tags			Spots
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			spot	body		requests.CreateSpotRequest	true	"Spot payload"
//	@Success		201		{object}	responses.SpotResponse
//	@Failure		400		{object}	apperror.ErrorResponse	"Bad Request"
//	@Failure		401		{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		500		{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/spots [post]
func (h *SpotHandler) Create(c *gin.Context) {
	// JWT Claims
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	// Request data
	var req requests.CreateSpotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidRequest)
		return
	}

	result, err := h.spotService.Create(c.Request.Context(), &req, userID)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// PATCH /api/v1/spots/:id
// UpdateSpot godoc
//
//	@Summary		Update a spot
//	@Description	Update spot details by ID. Only the owner or an admin can update a spot.
//	@Tags			Spots
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int							true	"Spot ID"
//	@Param			spot	body		requests.UpdateSpotRequest	true	"Spot update payload"
//	@Success		200		{object}	responses.SpotResponse
//	@Failure		400		{object}	apperror.ErrorResponse	"Bad Request"
//	@Failure		401		{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		403		{object}	apperror.ErrorResponse	"Forbidden - not owner or admin"
//	@Failure		404		{object}	apperror.ErrorResponse	"Spot not found"
//	@Failure		500		{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/spots/{id} [patch]
func (h *SpotHandler) Update(c *gin.Context) {
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

	var req requests.UpdateSpotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidRequest)
		return
	}

	result, err := h.spotService.Update(c.Request.Context(), uint(id), &req, userID, isAdmin)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

// DELETE /api/v1/spots/:id
// DeleteSpot godoc
//
//	@Summary		Delete a spot
//	@Description	Delete a spot by ID. Only the owner or an admin can delete a spot.
//	@Tags			Spots
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Spot ID"
//	@Success		200	{object}	nil	"Successfully deleted"
//	@Failure		400	{object}	apperror.ErrorResponse	"Invalid spot ID"
//	@Failure		401	{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		403	{object}	apperror.ErrorResponse	"Forbidden - not owner or admin"
//	@Failure		404	{object}	apperror.ErrorResponse	"Spot not found"
//	@Failure		500	{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/spots/{id} [delete]
func (h *SpotHandler) Delete(c *gin.Context) {
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

	err = h.spotService.Delete(c.Request.Context(), uint(id), userID, isAdmin)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}
