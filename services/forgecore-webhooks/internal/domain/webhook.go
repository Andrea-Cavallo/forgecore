package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	DeliveryStatusPending   = "pending"
	DeliveryStatusDelivered = "delivered"
	DeliveryStatusFailed    = "failed"

	MaxDeliveryAttempts = 5
)

type WebhookEndpoint struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	URL       string
	Secret    string
	Events    []string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (e *WebhookEndpoint) ListensTo(eventType string) bool {
	for _, ev := range e.Events {
		if ev == eventType || ev == "*" {
			return true
		}
	}
	return false
}

type WebhookDelivery struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	EndpointID  uuid.UUID
	EventType   string
	Payload     []byte
	Status      string
	Attempts    int
	LastError   string
	DeliveredAt *time.Time
	NextRetryAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (d *WebhookDelivery) CanRetry() bool {
	return d.Status == DeliveryStatusFailed && d.Attempts < MaxDeliveryAttempts
}
