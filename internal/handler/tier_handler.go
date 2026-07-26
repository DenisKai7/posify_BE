package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"posify-backend/internal/middleware"
	"posify-backend/internal/model"
	"posify-backend/internal/service"
)

type TierHandler struct {
	svc *service.TierService
}

func NewTierHandler(svc *service.TierService) *TierHandler {
	return &TierHandler{svc: svc}
}

// List handles GET /api/v1/tiers
func (h *TierHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	items, err := h.svc.List(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tiers": items})
}

// Update handles PUT /api/v1/tiers/:id
func (h *TierHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	tierID := chi.URLParam(r, "id")
	var req model.UpdateTierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	t, err := h.svc.Update(r.Context(), tenantID, tierID, req)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, t)
}
