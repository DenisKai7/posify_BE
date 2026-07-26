package handler

import (
	"encoding/json"
	"net/http"

	"posify-backend/internal/middleware"
	"posify-backend/internal/model"
	"posify-backend/internal/service"
)

type ShiftHandler struct {
	svc *service.ShiftService
}

func NewShiftHandler(svc *service.ShiftService) *ShiftHandler {
	return &ShiftHandler{svc: svc}
}

// Start handles POST /api/v1/shifts/start
func (h *ShiftHandler) Start(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	var req model.StartShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	shift, err := h.svc.Start(r.Context(), tenantID, userID, req.InitialCash)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, shift)
}

// Close handles POST /api/v1/shifts/close
func (h *ShiftHandler) Close(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	var req model.CloseShiftRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	shift, err := h.svc.Close(r.Context(), tenantID, userID, req.ActualCash)
	if err != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, shift)
}

// GetCurrent handles GET /api/v1/shifts/current
func (h *ShiftHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	tenantID, _ := r.Context().Value(middleware.TenantIDKey).(string)
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	shift, err := h.svc.GetCurrent(r.Context(), tenantID, userID)
	if err != nil {
		// No open shift is not an error — return null
		writeJSON(w, http.StatusOK, nil)
		return
	}
	writeJSON(w, http.StatusOK, shift)
}
