package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"posify-backend/internal/middleware"
	"posify-backend/internal/model"
	"posify-backend/internal/service"
)

type CustomerHandler struct {
	svc *service.CustomerService
}

func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

// Create handles POST /api/v1/customers
func (h *CustomerHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	var req model.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.Name == "" || req.Phone == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and phone required"})
		return
	}
	c, err := h.svc.Create(r.Context(), tenantID, req)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

// Get handles GET /api/v1/customers/:id
func (h *CustomerHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	customerID := chi.URLParam(r, "id")
	c, err := h.svc.GetByID(r.Context(), tenantID, customerID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// Search handles GET /api/v1/customers?search=...
func (h *CustomerHandler) Search(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	query := r.URL.Query().Get("search")
	if query == "" {
		// Return all customers
		items, err := h.svc.List(r.Context(), tenantID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"customers": items})
		return
	}
	items, err := h.svc.Search(r.Context(), tenantID, query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"customers": items})
}

// RedeemPoints handles POST /api/v1/customers/:id/redeem
func (h *CustomerHandler) RedeemPoints(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	customerID := chi.URLParam(r, "id")
	var req model.RedeemPointsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if err := h.svc.RedeemPoints(r.Context(), tenantID, customerID, req.Points); err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "redeemed"})
}
