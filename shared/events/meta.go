package events

import (
	"time"

	"github.com/google/uuid"
)

const EventVersionV1 = 1

type Metadata struct {
	Version       int       `json:"version"`
	EventName     string    `json:"event_name"`
	TenantID      uuid.UUID `json:"tenant_id"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func NewMetadata(eventName string, tenantID uuid.UUID, correlationID string, occurredAt time.Time) Metadata {
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	return Metadata{
		Version:       EventVersionV1,
		EventName:     eventName,
		TenantID:      tenantID,
		CorrelationID: correlationID,
		OccurredAt:    occurredAt,
	}
}

type Versioned interface {
	EventMetadata() Metadata
}
