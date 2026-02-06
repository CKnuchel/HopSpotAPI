package handler

import (
	"net/http"
	"strconv"

	"hopSpotAPI/internal/domain"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/service"
	"hopSpotAPI/pkg/apperror"
	"hopSpotAPI/pkg/logger"

	"github.com/gin-gonic/gin"
)

type PhotoHandler struct {
	photoService service.PhotoService
}

func NewPhotoHandler(photoService service.PhotoService) *PhotoHandler {
	return &PhotoHandler{photoService: photoService}
}

// POST /api/v1/spots/:id/photos
// godoc
//
//	@Summary		Upload a photo for a spot
//	@Description	Uploads a photo to the specified spot
//	@Tags			Photos
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id		path		int		true	"Spot ID"
//
//	@Param			photo	formData	file	true	"Photo file"
//	@Param			is_main	formData	bool	false	"Als Hauptbild setzen"
//
//	@Success		201		{object}	responses.PhotoResponse
//	@Failure		400		{object}	apperror.ErrorResponse
//	@Failure		401		{object}	apperror.ErrorResponse
//	@Failure		404		{object}	apperror.ErrorResponse
//	@Failure		500		{object}	apperror.ErrorResponse
//	@Router			/api/v1/spots/{id}/photos [post]
func (h *PhotoHandler) Upload(c *gin.Context) {
	// JWT Claims
	userID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		apperror.RespondWithError(c, apperror.AppErrSystemInternal)
		return
	}

	// Spot ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	// Photo file from form data
	file, err := c.FormFile("photo")
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationFieldRequired.WithDetails("Photo file is required"))
		return
	}

	// Is main photo
	isMain := c.PostForm("is_main") == "true"

	result, err := h.photoService.Upload(c.Request.Context(), uint(id), userID, file, isMain)
	if err != nil {
		logger.Log.Error().Err(err).Msg("Photo upload failed")
		apperror.RespondWithMappedError(c, err)
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
//	@Failure		400	{object}	apperror.ErrorResponse
//	@Failure		401	{object}	apperror.ErrorResponse
//	@Failure		403	{object}	apperror.ErrorResponse
//	@Failure		404	{object}	apperror.ErrorResponse
//	@Failure		500	{object}	apperror.ErrorResponse
//	@Router			/api/v1/photos/{id} [delete]
func (h *PhotoHandler) Delete(c *gin.Context) {
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

	photoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	err = h.photoService.Delete(c.Request.Context(), uint(photoID), userID, isAdmin)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// PATCH /api/v1/photos/:id/main
// godoc
//
//	@Summary		Set main photo
//	@Description	Sets a photo as the main photo for its spot
//	@Tags			Photos
//	@Param			id	path	int	true	"Photo ID"
//
//	@Success		204
//	@Failure		400	{object}	apperror.ErrorResponse
//	@Failure		401	{object}	apperror.ErrorResponse
//	@Failure		403	{object}	apperror.ErrorResponse
//	@Failure		404	{object}	apperror.ErrorResponse
//	@Failure		500	{object}	apperror.ErrorResponse
//	@Router			/api/v1/photos/{id}/main [patch]
func (h *PhotoHandler) SetMainPhoto(c *gin.Context) {
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

	photoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}
	err = h.photoService.SetMainPhoto(c.Request.Context(), uint(photoID), userID, isAdmin)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GET /api/v1/spots/:id/photos
// godoc
//
//	@Summary		Get photos by spot ID
//	@Description	Retrieves all photos for a specific spot
//	@Tags			Photos
//	@Param			id	path	int	true	"Spot ID"
//
//	@Success		200	{array}		responses.PhotoResponse
//	@Failure		400	{object}	apperror.ErrorResponse
//	@Failure		500	{object}	apperror.ErrorResponse
//	@Router			/api/v1/spots/{id}/photos [get]
func (h *PhotoHandler) GetBySpotID(c *gin.Context) {
	spotID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}
	resp, err := h.photoService.GetBySpotID(c.Request.Context(), uint(spotID))
	if err != nil {
		apperror.RespondWithMappedError(c, err)
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
//	@Failure		400		{object}	apperror.ErrorResponse
//	@Failure		404		{object}	apperror.ErrorResponse
//	@Failure		500		{object}	apperror.ErrorResponse
//	@Router			/api/v1/photos/{id}/url [get]
func (h *PhotoHandler) GetPresignedURL(c *gin.Context) {
	photoID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apperror.RespondWithError(c, apperror.AppErrValidationInvalidID)
		return
	}

	size := c.Query("size")

	url, err := h.photoService.GetPresignedURL(c.Request.Context(), uint(photoID), size)
	if err != nil {
		apperror.RespondWithMappedError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url})
}
