package domain

import (
	"time"

	"github.com/google/uuid"
)

// Well-known config keys used across services.
const (
	KeyMFARequired           = "mfa.required"
	KeyPasswordMinLen        = "password.min_length"
	KeySessionTTLHours       = "session.ttl_hours"
	KeyAllowedOAuthProviders = "oauth.providers"

	// CacheTTLSeconds is the default TTL for config values cached in Redis.
	CacheTTLSeconds = 300
)

type TenantConfig struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Key       string
	Value     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FeatureFlag struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Name      string
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
