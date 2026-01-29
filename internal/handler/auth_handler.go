package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

 "hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// POST /api/v1/auth/register
// Register godoc
//
//	@Summary		Register a new user
//	@Description	Registers a new user and returns access and refresh tokens
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			registerRequest	body		requests.RegisterRequest	true	"Register Request"
//	@Success		201				{object}	responses.LoginResponse
//	@Failure		400				{object}	map[string]string
//	@Failure		409				{object}	map[string]string
//	@Router			/api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req requests.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// POST /api/v1/auth/login
// Login godoc
//
//	@Summary		Login a user
//	@Description	Authenticates a user and returns access and refresh tokens
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			loginRequest	body		requests.LoginRequest	true	"Login Request"
//	@Success		200				{object}	responses.LoginResponse
//	@Failure		400				{object}	map[string]string
//	@Failure		401				{object}	map[string]string
//	@Failure		403				{object}	map[string]string
//	@Router			/api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req requests.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// POST /api/v1/auth/refresh-fcm-token
// RefreshFCMToken godoc
//
//	@Summary		Refresh FCM Token
//	@Description	Updates the FCM token for the authenticated user
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			refreshFCMTokenRequest	body		requests.RefreshFCMTokenRequest	true	"Refresh FCM Token Request"
//	@Success		200						{object}	map[string]string
//	@Failure		400						{object}	map[string]string
//	@Failure		401						{object}	map[string]string
//	@Router			/auth/refresh-fcm-token [post]
func (h *AuthHandler) RefreshFCMToken(c *gin.Context) {
	// Get UserId from context (set by auth middleware)
	userId, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Bind request body
	var req requests.RefreshFCMTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Call service to refresh FCM token
	userID, ok := userId.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}
	err := h.authService.RefreshFCMToken(c.Request.Context(), userID, req.FCMToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "FCM token updated successfully"})
}

// POST /api/v1/auth/refresh
// Refresh godoc
//
//	@Summary		Refresh tokens
//	@Description	Generates new access and refresh tokens using a valid refresh token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			refreshRequest	body		requests.RefreshTokenRequest	true	"Refresh Token Request"
//	@Success		200				{object}	responses.LoginResponse
//	@Failure		400				{object}	map[string]string
//	@Failure		401				{object}	map[string]string
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req requests.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result, err := h.authService.Refresh(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// POST /api/v1/auth/logout
// Logout godoc
//
//	@Summary		Logout user
//	@Description	Invalidates the refresh token
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			logoutRequest	body	requests.LogoutRequest	true	"Logout Request"
//	@Success		204				"No Content"
//	@Failure		400				{object}	map[string]string
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req requests.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.authService.Logout(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
