package handler

import (
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
//	@Param			id		path		int				true	"Bench ID"
//	@Param			photos	formData	file			true	"Photo file to upload"
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
// Löscht ein Foto (nur Uploader oder Admin)
func (h *PhotoHandler) Delete(c *gin.Context) {
	panic("TODO: implement")
}

// PATCH /api/v1/photos/:id/main
// Setzt ein Foto als Hauptbild (nur Bench-Ersteller oder Admin)
func (h *PhotoHandler) SetMainPhoto(c *gin.Context) {
	panic("TODO: implement")
}

// GET /api/v1/benches/:id/photos
// Gibt alle Fotos einer Bank zurück
func (h *PhotoHandler) GetByBenchID(c *gin.Context) {
	panic("TODO: implement")
}

// GET /api/v1/photos/:id/url
// Gibt eine temporäre URL für ein Foto zurück
// Query-Param: size (original, medium, thumbnail)
func (h *PhotoHandler) GetPresignedURL(c *gin.Context) {
	panic("TODO: implement")
}
