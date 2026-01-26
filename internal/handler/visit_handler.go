package handler

import "hopSpotAPI/internal/service"

type VisitHandler struct {
	visitService service.VisitService
}

func NewVisitHandler(visitService service.VisitService) *VisitHandler {
	return &VisitHandler{visitService: visitService}
}

// GET /api/v1/benches/:id/visits/count -- TODO: implement

// GET /api/v1/visits -- TODO: implement

// POST /api/v1/visits -- TODO: implement
