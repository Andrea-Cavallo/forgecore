// Package rest exposes the permission-service HTTP API.
package rest

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/permission-service/internal/application"
)

const (
	headerTenantID = "X-Tenant-ID"
	headerUserID   = "X-User-ID"
)

// Handler holds all HTTP handlers for the permission service.
type Handler struct {
	check *application.CheckPermissionUseCase
	grant *application.GrantPermissionUseCase
}

// NewHandler constructs the permission REST handler.
func NewHandler(check *application.CheckPermissionUseCase, grant *application.GrantPermissionUseCase) *Handler {
	return &Handler{check: check, grant: grant}
}

// RegisterRoutes mounts permission routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/permissions/check", h.checkHandler)
	mux.HandleFunc("POST /v1/permissions/grant", h.grantHandler)
	mux.HandleFunc("DELETE /v1/permissions/{id}", h.revokeHandler)
}

func (h *Handler) checkHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, ok := extractIDs(w, r)
	if !ok {
		return
	}
	var body struct {
		ResourceType string     `json:"resource_type"`
		ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
		Action       string     `json:"action"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	out, err := h.check.Execute(r.Context(), application.CheckPermissionInput{
		TenantID:     tenantID,
		UserID:       userID,
		ResourceType: body.ResourceType,
		ResourceID:   body.ResourceID,
		Action:       body.Action,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) grantHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := extractIDs(w, r)
	if !ok {
		return
	}
	var body struct {
		UserID       uuid.UUID  `json:"user_id"`
		ResourceType string     `json:"resource_type"`
		ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
		Action       string     `json:"action"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if err := h.grant.Execute(r.Context(), application.GrantPermissionInput{
		TenantID:     tenantID,
		UserID:       body.UserID,
		ResourceType: body.ResourceType,
		ResourceID:   body.ResourceID,
		Action:       body.Action,
	}); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) revokeHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := extractIDs(w, r)
	if !ok {
		return
	}
	permID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "permission id non valido")
		return
	}
	_ = permID
	_ = tenantID
	// Revoke not yet exposed as a use case — placeholder returns 204.
	w.WriteHeader(http.StatusNoContent)
}

// ---- helpers ----

func extractIDs(w http.ResponseWriter, r *http.Request) (tenantID, userID uuid.UUID, ok bool) {
	tenantID, err := uuid.Parse(r.Header.Get(headerTenantID))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "tenant_id mancante o non valido")
		return uuid.Nil, uuid.Nil, false
	}
	userID, err = uuid.Parse(r.Header.Get(headerUserID))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "user_id mancante o non valido")
		return uuid.Nil, uuid.Nil, false
	}
	return tenantID, userID, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "corpo richiesta non valido")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
