package handler

import (
	"hopSpotAPI/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// GET /api/v1/admin/users
func (h *AdminHandler) ListUsers(c *gin.Context) {
	panic("unimplemented")
}

// PATCH /api/v1/admin/users/:id
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	panic("unimplemented")
}

// DELETE /api/v1/admin/users/:id
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	panic("unimplemented")
}

// GET /api/v1/admin/invitation-codes
func (h *AdminHandler) ListInvitationCodes(c *gin.Context) {
	panic("unimplemented")
}

// POST /api/v1/admin/invitation-codes
func (h *AdminHandler) CreateInvitationCode(c *gin.Context) {
	panic("unimplemented")
}
