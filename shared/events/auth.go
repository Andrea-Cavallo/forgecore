package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	SubjectUserRegistered  = "auth.user.registered"
	SubjectUserLogin       = "auth.user.login"
	SubjectPasswordReset   = "auth.user.password_reset"
	SubjectPasswordChanged = "auth.user.password_changed"
	SubjectMFAEnabled      = "auth.mfa.enabled"
	SubjectEmailVerified   = "auth.email.verified"
)

// UserRegistered is published when a new user completes registration.
type UserRegistered struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	VerifyURL  string    `json:"verify_url"`
	OccurredAt time.Time `json:"occurred_at"`
}

// UserLogin is published on every successful login.
type UserLogin struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	UserID     uuid.UUID `json:"user_id"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	OccurredAt time.Time `json:"occurred_at"`
}

// PasswordReset is published when a password reset is requested.
type PasswordReset struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	UserID     uuid.UUID `json:"user_id"`
	Email      string    `json:"email"`
	ResetURL   string    `json:"reset_url"`
	OccurredAt time.Time `json:"occurred_at"`
}
