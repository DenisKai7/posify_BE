package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"posify-backend/internal/middleware"
	"posify-backend/internal/model"
	"posify-backend/internal/service"
)

type StockHandler struct {
	svc *service.StockService
}

func NewStockHandler(svc *service.StockService) *StockHandler {
	return &StockHandler{svc: svc}
}

// Adjust handles POST /api/v1/inventory/adjust
func (h *StockHandler) Adjust(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	var req model.CreateStockAdjustmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	if req.ProductID == "" || req.Type == "" || req.QtyChange == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "product_id, type, qty_change required"})
		return
	}

	sa, err := h.svc.Adjust(r.Context(), tenantID, userID, req)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, sa)
}

// LowStockAlerts handles GET /api/v1/inventory/low-stock-alerts?threshold=10
func (h *StockHandler) LowStockAlerts(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	threshold := 10
	if t := r.URL.Query().Get("threshold"); t != "" {
		if v, err := strconv.Atoi(t); err == nil && v > 0 {
			threshold = v
		}
	}
	items, err := h.svc.LowStockAlerts(r.Context(), tenantID, threshold)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"alerts": items, "threshold": threshold})
}

// History handles GET /api/v1/inventory/history/:productId?limit=50
func (h *StockHandler) History(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	productID := chi.URLParam(r, "productId")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	items, err := h.svc.History(r.Context(), tenantID, productID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"history": items})
}
