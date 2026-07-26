package handler

import (
	"encoding/json"
	"net/http"
	"time"

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
	userID, _ := r.Context().Value(middleware.UserIDKey).(string)

	var req model.SyncPushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if len(req.Transactions) == 0 {
		writeJSON(w, http.StatusOK, model.SyncPushResponse{SyncedIDs: []string{}})
		return
	}

	resp, err := h.svc.Push(r.Context(), tenantID, userID, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// Pull handles GET /api/v1/sync/pull
// Supports ?last_synced_at=ISO8601 for delta sync (LWW)
func (h *SyncHandler) Pull(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := r.Context().Value(middleware.TenantIDKey).(string)
	if !ok || tenantID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing tenant"})
		return
	}

	var since *time.Time
	if raw := r.URL.Query().Get("last_synced_at"); raw != "" {
		t, err := time.Parse(time.RFC3339, raw)
		if err == nil {
			since = &t
		}
	}

	resp, err := h.svc.Pull(r.Context(), tenantID, since)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
