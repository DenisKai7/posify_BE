package handler

import (
	"encoding/json"
	"net/http"

	"posify-backend/internal/middleware"
	"posify-backend/internal/model"
	"posify-backend/internal/service"
)

type SyncHandler struct {
	svc *service.SyncService
}

func NewSyncHandler(svc *service.SyncService) *SyncHandler {
	return &SyncHandler{svc: svc}
}

// Push handles POST /api/v1/sync/push
func (h *SyncHandler) Push(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middleware.TenantIDKey).(string)
	if !ok || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing tenant"})
		return
	}

	var req model.SyncPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(req.Transactions) == 0 {
		writeJSON(w, http.StatusOK, model.SyncPushResponse{SyncedIDs: []string{}})
		return
	}

	resp, err := h.svc.Push(r.Context(), tenantID, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Pull handles GET /api/v1/sync/pull
func (h *SyncHandler) Pull(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middleware.TenantIDKey).(string)
	if !ok || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing tenant"})
		return
	}

	resp, err := h.svc.Pull(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
