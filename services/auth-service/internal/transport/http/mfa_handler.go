package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/auth-service/internal/application"
	"github.com/yourorg/golang-modules/shared/middleware"
)

// mfaEnableExecutor abstracts EnableMFAUseCase for testability.
type mfaEnableExecutor interface {
	Execute(ctx context.Context, in application.EnableMFAInput) (*application.EnableMFAOutput, error)
}

// mfaVerifyExecutor abstracts VerifyMFAUseCase for testability.
type mfaVerifyExecutor interface {
	Execute(ctx context.Context, in application.VerifyMFAInput) (*application.VerifyMFAOutput, error)
}

// mfaDisableExecutor abstracts DisableMFAUseCase for testability.
type mfaDisableExecutor interface {
	Execute(ctx context.Context, tenantID, userID uuid.UUID) error
}

// MFAHandler handles MFA-related HTTP endpoints.
type MFAHandler struct {
	enable  mfaEnableExecutor
	verify  mfaVerifyExecutor
	disable mfaDisableExecutor
}

func NewMFAHandler(enable mfaEnableExecutor, verify mfaVerifyExecutor, disable mfaDisableExecutor) *MFAHandler {
	return &MFAHandler{enable: enable, verify: verify, disable: disable}
}

// EnableMFA handles POST /v1/auth/mfa/enable — starts MFA setup and returns the TOTP secret.
func (h *MFAHandler) EnableMFA(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, ok := extractTenantAndUser(r)
	if !ok {
		writeMFAError(w, http.StatusUnauthorized, "tenant o user ID mancante")
		return
	}

	out, err := h.enable.Execute(r.Context(), application.EnableMFAInput{
		TenantID: tenantID,
		UserID:   userID,
	})
	if err != nil {
		if errors.Is(err, application.ErrMFAAlreadyEnabled) {
			writeMFAError(w, http.StatusConflict, err.Error())
			return
		}
		writeMFAError(w, http.StatusInternalServerError, "errore interno")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"secret":       out.Secret,
		"qr_code_url":  out.QRCodeURL,
		"backup_codes": out.BackupCodes,
	})
}

// VerifyMFA handles POST /v1/auth/mfa/verify — validates the TOTP code.
func (h *MFAHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, ok := extractTenantAndUser(r)
	if !ok {
		writeMFAError(w, http.StatusUnauthorized, "tenant o user ID mancante")
		return
	}

	var body struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeMFAError(w, http.StatusBadRequest, "corpo richiesta non valido")
		return
	}
	if body.Code == "" {
		writeMFAError(w, http.StatusUnprocessableEntity, "il campo 'code' è obbligatorio")
		return
	}

	out, err := h.verify.Execute(r.Context(), application.VerifyMFAInput{
		TenantID: tenantID,
		UserID:   userID,
		Code:     body.Code,
	})
	if err != nil {
		if errors.Is(err, application.ErrMFAInvalidCode) {
			writeMFAError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if errors.Is(err, application.ErrMFANotEnabled) {
			writeMFAError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeMFAError(w, http.StatusInternalServerError, "errore interno")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  out.Tokens.AccessToken,
		"refresh_token": out.Tokens.RefreshToken,
		"expires_in":    out.Tokens.ExpiresIn,
	})
}

// DisableMFA handles DELETE /v1/auth/mfa — disables MFA for the user.
func (h *MFAHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	tenantID, userID, ok := extractTenantAndUser(r)
	if !ok {
		writeMFAError(w, http.StatusUnauthorized, "tenant o user ID mancante")
		return
	}

	if err := h.disable.Execute(r.Context(), tenantID, userID); err != nil {
		if errors.Is(err, application.ErrMFANotEnabled) {
			writeMFAError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeMFAError(w, http.StatusInternalServerError, "errore interno")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func extractTenantAndUser(r *http.Request) (tenantID, userID uuid.UUID, ok bool) {
	tenantID, tenantOK := middleware.TenantFromContext(r.Context())
	userIDStr := r.Header.Get("X-User-ID")
	if !tenantOK || userIDStr == "" {
		return uuid.UUID{}, uuid.UUID{}, false
	}
	var err error
	userID, err = uuid.Parse(userIDStr)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, false
	}
	return tenantID, userID, true
}

func writeMFAError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
