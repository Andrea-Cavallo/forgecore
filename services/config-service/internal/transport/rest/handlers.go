// Package rest exposes the config-service HTTP API.
package rest

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/config-service/internal/application"
)

const headerTenantID = "X-Tenant-ID"

// Handler holds all HTTP handlers for the config service.
type Handler struct {
	get *application.GetConfigUseCase
	set *application.SetConfigUseCase
}

// NewHandler constructs the config REST handler.
func NewHandler(get *application.GetConfigUseCase, set *application.SetConfigUseCase) *Handler {
	return &Handler{get: get, set: set}
}

// RegisterRoutes mounts config routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/config/{key}", h.getHandler)
	mux.HandleFunc("PUT /v1/config/{key}", h.setHandler)
}

func (h *Handler) getHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := extractTenantID(w, r)
	if !ok {
		return
	}
	out, err := h.get.Execute(r.Context(), application.GetConfigInput{
		TenantID: tenantID,
		Key:      r.PathValue("key"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !out.Found {
		writeError(w, http.StatusNotFound, "configurazione non trovata")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"key": r.PathValue("key"), "value": out.Value})
}

func (h *Handler) setHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := extractTenantID(w, r)
	if !ok {
		return
	}
	var body struct {
		Value string `json:"value"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if err := h.set.Execute(r.Context(), application.SetConfigInput{
		TenantID: tenantID,
		Key:      r.PathValue("key"),
		Value:    body.Value,
	}); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- helpers ----

func extractTenantID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.Header.Get(headerTenantID))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "tenant_id mancante o non valido")
		return uuid.Nil, false
	}
	return id, true
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
