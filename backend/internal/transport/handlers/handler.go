package handlers

import (
	"backend/internal/infrastructure/ws"
	"backend/internal/usecase"
)

// Handler holds all use case dependencies for HTTP handlers.
type Handler struct {
	healthUC usecase.HealthUseCase
	verifier usecase.FirebaseTokenVerifier // nil disables auth (dev only)
	hub      *ws.Hub
}

// NewHandler constructs a Handler with all required use cases.
func NewHandler(healthUC usecase.HealthUseCase, verifier usecase.FirebaseTokenVerifier, hub *ws.Hub) *Handler {
	return &Handler{healthUC: healthUC, verifier: verifier, hub: hub}
}
