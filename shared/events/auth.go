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

	EventUserRegistered  = "auth.user.registered.v1"
	EventUserLogin       = "auth.user.login.v1"
	EventPasswordReset   = "auth.user.password_reset.v1"
	EventPasswordChanged = "auth.user.password_changed.v1"
	EventMFAEnabled      = "auth.mfa.enabled.v1"
	EventEmailVerified   = "auth.email.verified.v1"
)

// UserRegistered is published when a new user completes registration.
type UserRegistered struct {
	Version       int       `json:"version,omitempty"`
	EventName     string    `json:"event_name,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	TenantID      uuid.UUID `json:"tenant_id"`
	UserID        uuid.UUID `json:"user_id"`
	Email         string    `json:"email"`
	VerifyURL     string    `json:"verify_url"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e UserRegistered) EventMetadata() Metadata {
	return NewMetadata(EventUserRegistered, e.TenantID, e.CorrelationID, e.OccurredAt)
}

// UserLogin is published on every successful login.
type UserLogin struct {
	Version       int       `json:"version,omitempty"`
	EventName     string    `json:"event_name,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	TenantID      uuid.UUID `json:"tenant_id"`
	UserID        uuid.UUID `json:"user_id"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e UserLogin) EventMetadata() Metadata {
	return NewMetadata(EventUserLogin, e.TenantID, e.CorrelationID, e.OccurredAt)
}

// PasswordReset is published when a password reset is requested.
type PasswordReset struct {
	Version       int       `json:"version,omitempty"`
	EventName     string    `json:"event_name,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	TenantID      uuid.UUID `json:"tenant_id"`
	UserID        uuid.UUID `json:"user_id"`
	Email         string    `json:"email"`
	ResetURL      string    `json:"reset_url"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e PasswordReset) EventMetadata() Metadata {
	return NewMetadata(EventPasswordReset, e.TenantID, e.CorrelationID, e.OccurredAt)
}
