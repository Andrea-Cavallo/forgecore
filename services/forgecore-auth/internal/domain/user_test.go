package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-auth/internal/domain"
)

func TestUser_HasRole(t *testing.T) {
	tests := []struct {
		name  string
		roles []string
		check string
		want  bool
	}{
		{"ha ruolo user", []string{"user"}, "user", true},
		{"non ha ruolo admin", []string{"user"}, "admin", false},
		{"ha ruolo admin tra più ruoli", []string{"user", "admin"}, "admin", true},
		{"lista vuota", []string{}, "user", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &domain.User{ID: uuid.New(), Roles: tc.roles}
			if got := u.HasRole(tc.check); got != tc.want {
				t.Errorf("HasRole(%q) = %v, want %v", tc.check, got, tc.want)
			}
		})
	}
}

func TestUser_IsDeleted(t *testing.T) {
	u := &domain.User{ID: uuid.New()}
	if u.IsDeleted() {
		t.Error("nuovo utente non deve essere eliminato")
	}
	now := time.Now()
	u.DeletedAt = &now
	if !u.IsDeleted() {
		t.Error("utente con DeletedAt deve essere eliminato")
	}
}

func TestUser_IsLocked(t *testing.T) {
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	t.Run("non bloccato se LockedUntil è nil", func(t *testing.T) {
		u := &domain.User{}
		if u.IsLocked(now) {
			t.Error("non deve essere bloccato senza LockedUntil")
		}
	})
	t.Run("non bloccato se LockedUntil è nel passato", func(t *testing.T) {
		u := &domain.User{LockedUntil: &past}
		if u.IsLocked(now) {
			t.Error("non deve essere bloccato se il tempo di blocco è scaduto")
		}
	})
	t.Run("bloccato se LockedUntil è nel futuro", func(t *testing.T) {
		u := &domain.User{LockedUntil: &future}
		if !u.IsLocked(now) {
			t.Error("deve essere bloccato se LockedUntil è nel futuro")
		}
	})
}

func TestTokenClaims_Fields(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	claims := domain.TokenClaims{
		UserID:      userID,
		TenantID:    tenantID,
		Roles:       []string{"user"},
		JTI:         "test-jti",
		MFAVerified: false,
	}
	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.TenantID != tenantID {
		t.Errorf("TenantID = %v, want %v", claims.TenantID, tenantID)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "user" {
		t.Errorf("Roles = %v, want [user]", claims.Roles)
	}
}

func TestErrors_Sentinel(t *testing.T) {
	errs := []error{
		domain.ErrUserNotFound,
		domain.ErrEmailAlreadyExists,
		domain.ErrInvalidCredentials,
		domain.ErrAccountLocked,
		domain.ErrTokenInvalid,
		domain.ErrTokenExpired,
		domain.ErrMFARequired,
		domain.ErrMFAInvalid,
		domain.ErrSessionNotFound,
	}
	for _, err := range errs {
		if err == nil {
			t.Error("sentinel error non deve essere nil")
		}
		if err.Error() == "" {
			t.Errorf("sentinel error deve avere un messaggio: %T", err)
		}
	}
}
