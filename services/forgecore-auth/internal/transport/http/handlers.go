package http

import (
	"encoding/json"
	"net/http"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-auth/internal/application"
	"github.com/Andrea-Cavallo/golang-modules/shared/middleware"
)

type Handler struct {
	register *application.RegisterUseCase
	login    *application.LoginUseCase
}

func NewHandler(register *application.RegisterUseCase, login *application.LoginUseCase) *Handler {
	return &Handler{register: register, login: login}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantFromContext(r.Context())
	if !ok {
		http.Error(w, "tenant mancante", http.StatusUnauthorized)
		return
	}
	var body struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "corpo richiesta non valido", http.StatusBadRequest)
		return
	}
	out, err := h.register.Execute(r.Context(), application.RegisterInput{
		TenantID:  tenantID,
		Email:     body.Email,
		Password:  body.Password,
		FirstName: body.FirstName,
		LastName:  body.LastName,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"user_id": out.UserID})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.TenantFromContext(r.Context())
	if !ok {
		http.Error(w, "tenant mancante", http.StatusUnauthorized)
		return
	}
	var body struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		DeviceID  string `json:"device_id"`
		UserAgent string `json:"user_agent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "corpo richiesta non valido", http.StatusBadRequest)
		return
	}
	out, err := h.login.Execute(r.Context(), application.LoginInput{
		TenantID:  tenantID,
		Email:     body.Email,
		Password:  body.Password,
		IPAddress: r.RemoteAddr,
		UserAgent: r.UserAgent(),
		DeviceID:  body.DeviceID,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  out.Tokens.AccessToken,
		"refresh_token": out.Tokens.RefreshToken,
		"expires_in":    out.Tokens.ExpiresIn,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, "errore serializzazione risposta", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnprocessableEntity)
	if encErr := json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}); encErr != nil {
		http.Error(w, "errore serializzazione errore", http.StatusInternalServerError)
	}
}
