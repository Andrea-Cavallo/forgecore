// Package rest exposes the forgecore-webhooks HTTP API.
package rest

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-webhooks/internal/application"
)

const headerTenantID = "X-Tenant-ID"

// Handler holds all HTTP handlers for the webhook service.
type Handler struct {
	register *application.RegisterEndpointUseCase
	deliver  *application.DeliverUseCase
}

// NewHandler constructs the webhook REST handler.
func NewHandler(register *application.RegisterEndpointUseCase, deliver *application.DeliverUseCase) *Handler {
	return &Handler{register: register, deliver: deliver}
}

// RegisterRoutes mounts webhook routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/webhooks/endpoints", h.registerHandler)
	mux.HandleFunc("POST /v1/webhooks/deliver", h.deliverHandler)
}

func (h *Handler) registerHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := extractTenantID(w, r)
	if !ok {
		return
	}
	var body struct {
		URL    string   `json:"url"`
		Secret string   `json:"secret"`
		Events []string `json:"events"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	out, err := h.register.Execute(r.Context(), application.RegisterEndpointInput{
		TenantID: tenantID,
		URL:      body.URL,
		Secret:   body.Secret,
		Events:   body.Events,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out.Endpoint)
}

func (h *Handler) deliverHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := extractTenantID(w, r)
	if !ok {
		return
	}
	var body struct {
		EventType string          `json:"event_type"`
		Payload   json.RawMessage `json:"payload"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	if err := h.deliver.Execute(r.Context(), application.DeliverInput{
		TenantID:  tenantID,
		EventType: body.EventType,
		Payload:   body.Payload,
	}); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusAccepted)
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
