package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/middleware"
	"hopSpotAPI/internal/service"
	"hopSpotAPI/pkg/logger"
)

type AdminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// GET /api/v1/admin/users
// ListUsers godoc
//
//	@Summary		List users
//	@Description	Get a paginated list of users with optional filters
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"				default(1)
//	@Param			limit		query		int		false	"Number of users per page"	default(50)
//	@Param			search		query		string	false	"Search term for username or email"
//	@Param			is_active	query		bool	false	"Filter by active status"
//	@Param			is_admin	query		bool	false	"Filter by admin status"
//	@Success		200			{object}	responses.PaginatedUsersResponse
//	@Failure		400
//	@Router			/api/v1/admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	var req requests.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}

	users, err := h.adminService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, users)
}

// PATCH /api/v1/admin/users/:id
// UpdateUser godoc
//
//	@Summary		Update a user
//	@Description	Update user details by ID
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		int								true	"User ID"
//	@Param			user	body		requests.AdminUpdateUserRequest	true	"User update payload"
//	@Success		200		{object}	responses.UserResponse
//	@Failure		400
//	@Router			/api/v1/admin/users/{id} [patch]
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req requests.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.adminService.UpdateUser(c.Request.Context(), uint(id), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DELETE /api/v1/admin/users/:id
// DeleteUser godoc
//
//	@Summary		Delete a user
//	@Description	Delete a user by ID
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"User ID"
//	@Success		204	"No Content"
//	@Failure		400
//	@Router			/api/v1/admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	adminID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.adminService.DeleteUser(c.Request.Context(), uint(id), adminID)
	if err != nil {
		logger.Error().Err(err).Uint("user_id", uint(id)).Msg("Failed to delete user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	// Audit Log
	logger.Info().
		Uint("admin_id", adminID).
		Uint("deleted_user_id", uint(id)).
		Msg("User deleted by admin")

	c.Status(http.StatusNoContent)
}

// GET /api/v1/admin/invitation-codes
// ListInvitationCodes godoc
//
//	@Summary		List invitation codes
//	@Description	Get a paginated list of invitation codes
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			page	query		int	false	"Page number"				default(1)
//	@Param			limit	query		int	false	"Number of codes per page"	default(50)
//	@Success		200		{object}	responses.PaginatedInvitationCodesResponse
//	@Failure		400
//	@Router			/api/v1/admin/invitation-codes [get]
func (h *AdminHandler) ListInvitationCodes(c *gin.Context) {
	var req = requests.ListInvitationCodesRequest{}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}

	result, err := h.adminService.ListInvitationCodes(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve invitation codes"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// POST /api/v1/admin/invitation-codes
// CreateInvitationCode godoc
//
//	@Summary		Create an invitation code
//	@Description	Create a new invitation code
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			invitation_code	body		requests.CreateInvitationCodeRequest	true	"Invitation code payload"
//	@Success		201				{object}	responses.InvitationCodeResponse
//	@Failure		400
//	@Router			/api/v1/admin/invitation-codes [post]
func (h *AdminHandler) CreateInvitationCode(c *gin.Context) {
	adminID, ok := c.MustGet(middleware.ContextKeyUserID).(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user context"})
		return
	}

	var req requests.CreateInvitationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.adminService.CreateInvitationCode(c.Request.Context(), &req, adminID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation code"})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// DELETE /api/v1/admin/invitation-codes/:id
// DeleteInvitationCode godoc
//
//	@Summary		Delete an invitation code
//	@Description	Delete an invitation code by ID (only if not redeemed)
//	@Tags			Admin
//	@Param			id	path	int	true	"Invitation Code ID"
//	@Success		204	"No Content"
//	@Failure		400
//	@Failure		404
//	@Router			/api/v1/admin/invitation-codes/{id} [delete]
func (h *AdminHandler) DeleteInvitationCode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation code ID"})
		return
	}

	err = h.adminService.DeleteInvitationCode(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error().Err(err).Uint("code_id", uint(id)).Msg("Failed to delete invitation code")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}