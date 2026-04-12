// Package rest exposes the subscription-service HTTP API.
package rest

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/subscription-service/internal/application"
)

const (
	headerTenantID = "X-Tenant-ID"
	headerUserID   = "X-User-ID"
)

// Handler holds all HTTP handlers for the subscription service.
type Handler struct {
	subscribe *application.SubscribeUseCase
	cancel    *application.CancelUseCase
}

// NewHandler constructs the subscription REST handler.
func NewHandler(subscribe *application.SubscribeUseCase, cancel *application.CancelUseCase) *Handler {
	return &Handler{subscribe: subscribe, cancel: cancel}
}

// RegisterRoutes mounts subscription routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/subscriptions", h.subscribeHandler)
	mux.HandleFunc("DELETE /v1/subscriptions/{id}", h.cancelHandler)
}

func (h *Handler) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, ok := extractIDs(w, r)
	if !ok {
		return
	}
	var body struct {
		PlanID     uuid.UUID `json:"plan_id"`
		CustomerID string    `json:"customer_id"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	out, err := h.subscribe.Execute(r.Context(), application.SubscribeInput{
		TenantID:   tenantID,
		UserID:     userID,
		PlanID:     body.PlanID,
		CustomerID: body.CustomerID,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out.Subscription)
}

func (h *Handler) cancelHandler(w http.ResponseWriter, r *http.Request) {
	tenantID, _, ok := extractIDs(w, r)
	if !ok {
		return
	}
	subID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "subscription id non valido")
		return
	}
	if err := h.cancel.Execute(r.Context(), application.CancelInput{
		SubscriptionID: subID,
		TenantID:       tenantID,
	}); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
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
