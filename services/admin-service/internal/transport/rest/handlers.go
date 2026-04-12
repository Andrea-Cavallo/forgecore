// Package rest exposes the admin-service HTTP API.
package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/admin-service/internal/application"
)

const headerTenantID = "X-Tenant-ID"

// Handler holds all HTTP handlers for the admin service.
type Handler struct {
	listTenants *application.ListTenantsUseCase
	manageUsers *application.ManageUsersUseCase
	stats       *application.GetStatsUseCase
}

// NewHandler constructs the admin REST handler.
func NewHandler(
	lt *application.ListTenantsUseCase,
	mu *application.ManageUsersUseCase,
	gs *application.GetStatsUseCase,
) *Handler {
	return &Handler{listTenants: lt, manageUsers: mu, stats: gs}
}

// RegisterRoutes mounts admin routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /v1/admin/tenants", h.listTenantsHandler)
	mux.HandleFunc("GET /v1/admin/users/{id}", h.getUserHandler)
	mux.HandleFunc("POST /v1/admin/users/{id}/disable", h.disableUserHandler)
	mux.HandleFunc("GET /v1/admin/stats", h.statsHandler)
}

func (h *Handler) listTenantsHandler(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	tenants, err := h.listTenants.Execute(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tenants": tenants})
}

func (h *Handler) getUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, tenantID, ok := parseUserAndTenant(w, r)
	if !ok {
		return
	}
	user, err := h.manageUsers.GetUser(r.Context(), userID, tenantID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) disableUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, tenantID, ok := parseUserAndTenant(w, r)
	if !ok {
		return
	}
	if err := h.manageUsers.DisableUser(r.Context(), userID, tenantID); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) statsHandler(w http.ResponseWriter, r *http.Request) {
	stats, err := h.stats.Execute(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// ---- helpers ----

func parseUserAndTenant(w http.ResponseWriter, r *http.Request) (userID, tenantID uuid.UUID, ok bool) {
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "user id non valido")
		return uuid.Nil, uuid.Nil, false
	}
	tenantID, err = uuid.Parse(r.Header.Get(headerTenantID))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "tenant_id mancante o non valido")
		return uuid.Nil, uuid.Nil, false
	}
	return userID, tenantID, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
