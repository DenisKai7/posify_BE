package handler

import (
	"encoding/json"
	"net/http"

	"posify-backend/internal/middleware"
	"posify-backend/internal/model"
	"posify-backend/internal/service"
)

type DiscountHandler struct {
	svc *service.DiscountService
}

func NewDiscountHandler(svc *service.DiscountService) *DiscountHandler {
	return &DiscountHandler{svc: svc}
}

// Create handles POST /api/v1/discounts
func (h *DiscountHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	var req model.CreateDiscountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.Code == "" || req.DiscountValue <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "code and discount_value required"})
		return
	}
	d, err := h.svc.Create(r.Context(), tenantID, req)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, d)
}

// List handles GET /api/v1/discounts
func (h *DiscountHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	items, err := h.svc.List(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"discounts": items})
}

// Validate handles POST /api/v1/discounts/validate
func (h *DiscountHandler) Validate(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	var req model.ValidateDiscountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	result, err := h.svc.Validate(r.Context(), tenantID, req.Code, req.Subtotal)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}
