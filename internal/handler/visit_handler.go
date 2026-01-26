package handler

import (
	"hopSpotAPI/internal/service"

	"github.com/gin-gonic/gin"
)

type VisitHandler struct {
	visitService service.VisitService
}

func NewVisitHandler(visitService service.VisitService) *VisitHandler {
	return &VisitHandler{visitService: visitService}
}

// GET /api/v1/benches/:id/visits/count -- TODO: implement
func (h *VisitHandler) GetVisitCountByBenchID(c *gin.Context) {
	panic("To be done")
}

// GET /api/v1/visits -- TODO: implement
func (h *VisitHandler) ListVisits(c *gin.Context) {
	panic("To be done")
}

// POST /api/v1/visits -- TODO: implement
func (h *VisitHandler) CreateVisit(c *gin.Context) {
	panic("To be done")
}
