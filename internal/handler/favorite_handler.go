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

// POST /api/v1/benches/:id/favorite
// AddFavorite godoc
//
//	@Summary		Add bench to favorites
//	@Description	Add a bench to the user's favorites
//	@Tags			Favorites
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Bench ID"
//	@Success		201	{object}	map[string]bool	"is_favorite: true"
//	@Failure		400	{object}	apperror.ErrorResponse	"Invalid bench ID"
//	@Failure		401	{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		409	{object}	apperror.ErrorResponse	"Already in favorites"
//	@Failure		500	{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/benches/{id}/favorite [post]
func (h *FavoriteHandler) Add(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	benchID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	if err := h.favoriteService.Add(c.Request.Context(), userID, uint(benchID)); err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"is_favorite": true})
}

// DELETE /api/v1/benches/:id/favorite
// RemoveFavorite godoc
//
//	@Summary		Remove bench from favorites
//	@Description	Remove a bench from the user's favorites
//	@Tags			Favorites
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"Bench ID"
//	@Success		200	{object}	map[string]bool	"is_favorite: false"
//	@Failure		400	{object}	apperror.ErrorResponse	"Invalid bench ID"
//	@Failure		401	{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		404	{object}	apperror.ErrorResponse	"Favorite not found"
//	@Failure		500	{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/benches/{id}/favorite [delete]
func (h *FavoriteHandler) Remove(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	benchID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	if err := h.favoriteService.Remove(c.Request.Context(), userID, uint(benchID)); err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_favorite": false})
}

// GET /api/v1/benches/:id/favorite
// CheckFavorite godoc
//
//	@Summary		Check if bench is favorited
//	@Description	Check if a bench is in the user's favorites
//	@Tags			Favorites
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"Bench ID"
//	@Success		200	{object}	map[string]bool	"is_favorite: true/false"
//	@Failure		400	{object}	apperror.ErrorResponse	"Invalid bench ID"
//	@Failure		401	{object}	apperror.ErrorResponse	"Unauthorized"
//	@Failure		500	{object}	apperror.ErrorResponse	"Internal Server Error"
//	@Router			/api/v1/benches/{id}/favorite [get]
func (h *FavoriteHandler) Check(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	benchID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	isFavorite, err := h.favoriteService.IsFavorite(c.Request.Context(), userID, uint(benchID))
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
//	@Description	Get a paginated list of the user's favorite benches
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
