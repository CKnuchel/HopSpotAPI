package handler

import (
	"net/http"
	"strconv"

	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/service"
	"hopSpotAPI/pkg/apperror"

	"github.com/gin-gonic/gin"
)

type FavoriteHandler struct {
	favoriteService service.FavoriteService
}

func NewFavoriteHandler(favoriteService service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{favoriteService: favoriteService}
}

// POST /api/v1/spots/:id/favorite
// AddFavorite godoc
//
//	@Summary		Add spot to favorites
//	@Description	Add a spot to the user's favorites
//	@Tags			Favorites
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Spot ID"
//	@Success		201	{object}	map[string]bool	"is_favorite: true"
//	@Failure		400	{object}	apperror.ErrorResponse	"Invalid spot ID"
//	@Failure		401	{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		409	{object}	apperror.ErrorResponse	"Already in favorites"
//	@Failure		500	{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/spots/{id}/favorite [post]
func (h *FavoriteHandler) Add(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	spotID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	if err := h.favoriteService.Add(c.Request.Context(), userID, uint(spotID)); err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"is_favorite": true})
}

// DELETE /api/v1/spots/:id/favorite
// RemoveFavorite godoc
//
//	@Summary		Remove spot from favorites
//	@Description	Remove a spot from the user's favorites
//	@Tags			Favorites
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Spot ID"
//	@Success		200	{object}	map[string]bool	"is_favorite: false"
//	@Failure		400	{object}	apperror.ErrorResponse	"Invalid spot ID"
//	@Failure		401	{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		404	{object}	apperror.ErrorResponse	"Favorite not found"
//	@Failure		500	{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/spots/{id}/favorite [delete]
func (h *FavoriteHandler) Remove(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	spotID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	if err := h.favoriteService.Remove(c.Request.Context(), userID, uint(spotID)); err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_favorite": false})
}

// GET /api/v1/spots/:id/favorite
// CheckFavorite godoc
//
//	@Summary		Check if spot is favorited
//	@Description	Check if a spot is in the user's favorites
//	@Tags			Favorites
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Spot ID"
//	@Success		200	{object}	map[string]bool	"is_favorite: true/false"
//	@Failure		400	{object}	apperror.ErrorResponse	"Invalid spot ID"
//	@Failure		401	{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		500	{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/spots/{id}/favorite [get]
func (h *FavoriteHandler) Check(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	spotID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	isFavorite, err := h.favoriteService.IsFavorite(c.Request.Context(), userID, uint(spotID))
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_favorite": isFavorite})
}

// GET /api/v1/favorites
// ListFavorites godoc
//
//	@Summary		List user's favorites
//	@Description	Get a paginated list of the user's favorite spots
//	@Tags			Favorites
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int	false	"Page number"	default(1)
//	@Param			limit	query		int	false	"Items per page"	default(50)
//	@Success		200		{object}	responses.PaginatedFavoritesResponse
//	@Failure		401		{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		500		{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/favorites [get]
func (h *FavoriteHandler) List(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	response, err := h.favoriteService.List(c.Request.Context(), userID, page, limit)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
