package handler

import (
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GET /api/v1/users/me
// GetProfile godoc
// @Summary Get user profile
// @Description Retrieve the profile of the currently authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} responses.UserResponse
// @Failure 400
// @Router /api/v1/users/me [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userId := c.MustGet(middleware.ContextKeyUserID)

	result, err := h.userService.GetProfile(c.Request.Context(), userId.(uint))
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// PATCH /api/v1/users/me
// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the profile of the currently authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Param profile body requests.UpdateProfileRequest true "Profile payload"
// @Success 200 {object} responses.UserResponse
// @Failure 400
// @Router /api/v1/users/me [patch]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userId := c.MustGet(middleware.ContextKeyUserID)

	var req requests.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := h.userService.UpdateProfile(c.Request.Context(), userId.(uint), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// POST /api/v1/users/me/change-password
// ChangePassword godoc
// @Summary Change user password
// @Description Change the password of the currently authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Param password body requests.ChangePasswordRequest true "Change password payload"
// @Success 204 "No Content"
// @Failure 400
// @Router /api/v1/users/me/change-password [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userId := c.MustGet(middleware.ContextKeyUserID)

	var req requests.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.userService.ChangePassword(c.Request.Context(), userId.(uint), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
