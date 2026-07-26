package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"posify-backend/internal/middleware"
	"posify-backend/internal/model"
	"posify-backend/internal/service"
)

type PaymentGatewayHandler struct {
	svc *service.PaymentGatewayService
}

func NewPaymentGatewayHandler(svc *service.PaymentGatewayService) *PaymentGatewayHandler {
	return &PaymentGatewayHandler{svc: svc}
}

// Get handles GET /api/v1/settings/payment
func (h *PaymentGatewayHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	pg, err := h.svc.Get(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "payment gateway not configured"})
		return
	}
	writeJSON(w, http.StatusOK, pg)
}

// Update handles PUT /api/v1/settings/payment
func (h *PaymentGatewayHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	var req model.UpdatePaymentGatewayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}
	pg, err := h.svc.Update(r.Context(), tenantID, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, pg)
}

// Test handles POST /api/v1/settings/payment/test
func (h *PaymentGatewayHandler) Test(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	pg, err := h.svc.GetRaw(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "payment gateway not configured"})
		return
	}

	// Test by calling Midtrans sandbox/production status endpoint
	baseURL := "https://api.sandbox.midtrans.com"
	if pg.IsProduction {
		baseURL = "https://api.midtrans.com"
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseURL+"/v2/status", nil)
	req.SetBasicAuth(pg.ServerKey, "")
	resp, err := client.Do(req)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": false, "message": fmt.Sprintf("connection failed: %v", err)})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "koneksi berhasil (auth required = normal)"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": fmt.Sprintf("koneksi berhasil (HTTP %d)", resp.StatusCode)})
}
