package handler

import (
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// POST /api/v1/auth/register
// Register godoc
// @Summary      Register a new user
// @Description  Registers a new user with the provided details
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        registerRequest  body      requests.RegisterRequest  true  "Register Request"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /auth/register [post]
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
// @Summary      Login a user
// @Description  Authenticates a user and returns a JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        loginRequest  body      requests.LoginRequest  true  "Login Request"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]string
// @Router       /auth/login [post]
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
// @Summary      Refresh FCM Token
// @Description  Updates the FCM token for the authenticated user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        refreshFCMTokenRequest  body      requests.RefreshFCMTokenRequest  true  "Refresh FCM Token Request"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Router       /auth/refresh-fcm-token [post]
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
	err := h.authService.RefreshFCMToken(c.Request.Context(), userId.(uint), req.FCMToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "FCM token updated successfully"})
}
