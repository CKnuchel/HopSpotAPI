package handler

import (
	"hopSpotAPI/internal/dto/requests"
	"hopSpotAPI/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BenchHandler struct {
	benchService service.BenchService
}

func NewBenchHandler(benchService service.BenchService) *BenchHandler {
	return &BenchHandler{benchService: benchService}
}

// GET /api/v1/benches
func (h *BenchHandler) List(c *gin.Context) {
	var req requests.ListBenchesRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Define default pagination values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	// Call the service to get benches
	benches, err := h.benchService.List(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve benches"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": benches})
}
