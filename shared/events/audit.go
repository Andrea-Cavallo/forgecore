package events

import (
	"time"

	"github.com/google/uuid"
)

// SubjectAuditWildcard matches all audit subjects in NATS subscriptions.
const (
	SubjectAuditWildcard = "audit.>"
	EventAuditRecorded   = "audit.recorded.v1"
)

// AuditEvent is published by every service for security-sensitive actions.
type AuditEvent struct {
	Version       int            `json:"version,omitempty"`
	EventName     string         `json:"event_name,omitempty"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	TenantID      uuid.UUID      `json:"tenant_id"`
	ActorID       *uuid.UUID     `json:"actor_id,omitempty"`
	ActorType     string         `json:"actor_type"` // "user", "system", "admin"
	Action        string         `json:"action"`     // "user.login", "payment.succeeded"
	ResourceID    *uuid.UUID     `json:"resource_id,omitempty"`
	ResourceType  string         `json:"resource_type,omitempty"`
	IPAddress     string         `json:"ip_address,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
	OccurredAt    time.Time      `json:"occurred_at"`
}

func (e AuditEvent) EventMetadata() Metadata {
	return NewMetadata(EventAuditRecorded, e.TenantID, e.CorrelationID, e.OccurredAt)
}

// AuditSubject returns the NATS subject for a given action, e.g. "audit.user.login".
func AuditSubject(action string) string { return "audit." + action }
