package handler

import (
	"net/http"
	"strconv"

	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/service"

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
//	@Failure		400	{object}	map[string]string	"Invalid bench ID"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		500	{object}	map[string]string	"Internal Server Error"
//	@Router			/api/v1/benches/{id}/favorite [post]
func (h *FavoriteHandler) Add(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	benchID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bench ID"})
		return
	}

	if err := h.favoriteService.Add(c.Request.Context(), userID, uint(benchID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add favorite"})
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
//	@Failure		400	{object}	map[string]string	"Invalid bench ID"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		404	{object}	map[string]string	"Favorite not found"
//	@Failure		500	{object}	map[string]string	"Internal Server Error"
//	@Router			/api/v1/benches/{id}/favorite [delete]
func (h *FavoriteHandler) Remove(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	benchID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bench ID"})
		return
	}

	if err := h.favoriteService.Remove(c.Request.Context(), userID, uint(benchID)); err != nil {
		if err.Error() == "favorite not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Favorite not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove favorite"})
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
//	@Failure		400	{object}	map[string]string	"Invalid bench ID"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		500	{object}	map[string]string	"Internal Server Error"
//	@Router			/api/v1/benches/{id}/favorite [get]
func (h *FavoriteHandler) Check(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	benchID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bench ID"})
		return
	}

	isFavorite, err := h.favoriteService.IsFavorite(c.Request.Context(), userID, uint(benchID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check favorite"})
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
//	@Failure		401		{object}	map[string]string	"Unauthorized"
//	@Failure		500		{object}	map[string]string	"Internal Server Error"
//	@Router			/api/v1/favorites [get]
func (h *FavoriteHandler) List(c *gin.Context) {
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve favorites"})
		return
	}

	c.JSON(http.StatusOK, response)
}
