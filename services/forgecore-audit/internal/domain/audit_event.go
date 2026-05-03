package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	ActorTypeUser   = "user"
	ActorTypeSystem = "system"
	ActorTypeAdmin  = "admin"
)

// AuditEntry is immutable — never updated or deleted.
type AuditEntry struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	ActorID      *uuid.UUID
	ActorType    string
	Action       string
	ResourceID   *uuid.UUID
	ResourceType string
	IPAddress    string
	Metadata     map[string]any
	OccurredAt   time.Time
}
