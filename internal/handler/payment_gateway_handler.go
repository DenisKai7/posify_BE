package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"message": "Payment gateway belum dikonfigurasi",
			"data":    paymentGatewayData(nil, false),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Payment gateway ditemukan",
		"data":    paymentGatewayData(pg, true),
	})
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
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Pengaturan pembayaran disimpan",
		"data":    paymentGatewayData(pg, true),
	})
}

// Test handles POST /api/v1/settings/payment/test
func (h *PaymentGatewayHandler) Test(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	pg, err := h.svc.GetRaw(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"message": "Payment gateway belum dikonfigurasi. Klik Simpan terlebih dahulu.",
			"data":    paymentGatewayData(nil, false),
		})
		return
	}

	if pg.MerchantID == "" || pg.ClientKey == "" || pg.ServerKey == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"message": "Merchant ID, Client Key, dan Server Key wajib diisi lalu disimpan sebelum uji koneksi.",
			"data":    paymentGatewayData(pg, false),
		})
		return
	}

	baseURL := "https://api.sandbox.midtrans.com"
	if pg.IsProduction {
		baseURL = "https://api.midtrans.com"
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", baseURL+"/v2/status", nil)
	req.SetBasicAuth(pg.ServerKey, "")
	resp, err := client.Do(req)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"message": fmt.Sprintf("Koneksi ke Midtrans gagal: %v", err),
			"data":    paymentGatewayData(pg, true),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": false,
			"message": "Koneksi ke Midtrans gagal: Server Key tidak valid.",
			"data":    paymentGatewayData(pg, true),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"message": "Koneksi ke Midtrans berhasil!",
		"data":    paymentGatewayData(pg, true),
	})
}

func paymentGatewayData(pg *model.PaymentGateway, configured bool) map[string]any {
	if pg != nil {
		configured = configured && strings.TrimSpace(pg.MerchantID) != "" && strings.TrimSpace(pg.ClientKey) != "" && strings.TrimSpace(pg.ServerKey) != ""
	}

	if pg == nil {
		return map[string]any{
			"is_configured": false,
			"is_active":     false,
			"environment":   "sandbox",
			"provider":      "midtrans",
			"merchant_id":   "",
			"client_key":    "",
			"server_key":    "",
			"is_production": false,
			"webhook_url":   "/api/v1/payments/midtrans/webhook",
		}
	}

	environment := "sandbox"
	if pg.IsProduction {
		environment = "production"
	}

	return map[string]any{
		"is_configured": configured,
		"is_active":     pg.IsActive,
		"environment":   environment,
		"id":            pg.ID,
		"provider":      pg.Provider,
		"merchant_id":   pg.MerchantID,
		"client_key":    pg.ClientKey,
		"server_key":    pg.ServerKey,
		"is_production": pg.IsProduction,
		"webhook_url":   pg.WebhookURL,
	}
}
