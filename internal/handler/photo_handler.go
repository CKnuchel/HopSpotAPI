package handler

import (
	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PhotoHandler struct {
	photoService service.PhotoService
}

func NewPhotoHandler(photoService service.PhotoService) *PhotoHandler {
	return &PhotoHandler{photoService: photoService}
}

// POST /api/v1/benches/:id/photos
// godoc
//
//	@Summary		Upload a photo for a bench
//	@Description	Uploads a photo to the specified bench
//	@Tags			Photos
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id		path		int		true	"Bench ID"
//
//	@Param			photo	formData	file	true	"Photo file"
//	@Param			is_main	formData	bool	false	"Als Hauptbild setzen"
//
//	@Success		201		{object}	responses.PhotoResponse
//	@Failure		400
//	@Failure		401
//	@Failure		404
//	@Failure		500
//	@Router			/api/v1/benches/{id}/photos [post]
func (h *PhotoHandler) Upload(c *gin.Context) {
	// JWT Claims
	userID := c.MustGet(middleware.ContextKeyUserID).(uint)

	// Bench ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bench ID"})
		return
	}

	// Photo file from form data
	file, err := c.FormFile("photo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Photo file is required"})
		return
	}

	// Is main photo
	isMain := c.PostForm("is_main") == "true"

	result, err := h.photoService.Upload(c.Request.Context(), uint(id), userID, file, isMain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload photo"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// DELETE /api/v1/photos/:id
// godoc
//
//	@Summary		Delete a photo
//	@Description	Deletes a photo by its ID
//	@Tags			Photos
//	@Param			id	path	int	true	"Photo ID"
//
//	@Success		204
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		404
//	@Failure		500
//	@Router			/api/v1/photos/{id} [delete]
func (h *PhotoHandler) Delete(c *gin.Context) {
	// JWT Claims
	userID := c.MustGet(middleware.ContextKeyUserID).(uint)
	userRole := c.MustGet(middleware.ContextKeyUserRole).(domain.Role)
	isAdmin := userRole == domain.RoleAdmin

	photoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo ID"})
		return
	}

	err = h.photoService.Delete(c.Request.Context(), uint(photoID), userID, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete photo"})
		return
	}

	c.Status(http.StatusNoContent)
}

// PATCH /api/v1/photos/:id/main
// godoc
//
//	@Summary		Set main photo
//	@Description	Sets a photo as the main photo for its bench
//	@Tags			Photos
//	@Param			id	path	int	true	"Photo ID"
//
//	@Success		204
//	@Failure		400
//	@Failure		401
//	@Failure		403
//	@Failure		404
//	@Failure		500
//	@Router			/api/v1/photos/{id}/main [patch]
func (h *PhotoHandler) SetMainPhoto(c *gin.Context) {
	// JWT Claims
	userID := c.MustGet(middleware.ContextKeyUserID).(uint)
	userRole := c.MustGet(middleware.ContextKeyUserRole).(domain.Role)
	isAdmin := userRole == domain.RoleAdmin

	photoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo ID"})
		return
	}
	err = h.photoService.SetMainPhoto(c.Request.Context(), uint(photoID), userID, isAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set main photo"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /api/v1/benches/:id/photos
// godoc
//
//	@Summary		Get photos by bench ID
//	@Description	Retrieves all photos for a specific bench
//	@Tags			Photos
//	@Param			id	path	int	true	"Bench ID"
//
//	@Success		200	{array}	responses.PhotoResponse
//	@Failure		400
//	@Failure		500
//	@Router			/api/v1/benches/{id}/photos [get]
func (h *PhotoHandler) GetByBenchID(c *gin.Context) {
	benchID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bench ID"})
		return
	}
	resp, err := h.photoService.GetByBenchID(c.Request.Context(), uint(benchID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get photos"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GET /api/v1/photos/:id/url
// godoc
//
//	@Summary		Get presigned URL for photo
//	@Description	Retrieves a presigned URL for accessing a photo
//	@Tags			Photos
//	@Param			id		path		int		true	"Photo ID"
//	@Param			size	query		string	false	"Size of the photo (e.g., original, medium, thumbnail)"
//
//	@Success		200		{object}	map[string]string
//	@Failure		400
//	@Failure		404
//	@Failure		500
//	@Router			/api/v1/photos/{id}/url [get]
func (h *PhotoHandler) GetPresignedURL(c *gin.Context) {
	photoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo ID"})
		return
	}

	size := c.Query("size")

	url, err := h.photoService.GetPresignedURL(c.Request.Context(), uint(photoID), size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get presigned URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}
