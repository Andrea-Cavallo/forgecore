package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	SubjectNotificationRequested = "notification.requested"
	SubjectNotificationSent      = "notification.sent"
	SubjectNotificationFailed    = "notification.failed"

	EventNotificationRequested = "notification.requested.v1"
	EventNotificationSent      = "notification.sent.v1"
	EventNotificationFailed    = "notification.failed.v1"
)

// NotificationRequested is published when any service needs to send a notification.
type NotificationRequested struct {
	Version       int               `json:"version,omitempty"`
	EventName     string            `json:"event_name,omitempty"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	TenantID      uuid.UUID         `json:"tenant_id"`
	UserID        uuid.UUID         `json:"user_id"`
	Channel       string            `json:"channel"` // "email", "sms"
	Template      string            `json:"template"`
	Vars          map[string]string `json:"vars"`
	OccurredAt    time.Time         `json:"occurred_at"`
}

func (e NotificationRequested) EventMetadata() Metadata {
	return NewMetadata(EventNotificationRequested, e.TenantID, e.CorrelationID, e.OccurredAt)
}

// NotificationSent is published when a notification is successfully delivered.
type NotificationSent struct {
	Version        int       `json:"version,omitempty"`
	EventName      string    `json:"event_name,omitempty"`
	CorrelationID  string    `json:"correlation_id,omitempty"`
	TenantID       uuid.UUID `json:"tenant_id"`
	NotificationID uuid.UUID `json:"notification_id"`
	UserID         uuid.UUID `json:"user_id"`
	Channel        string    `json:"channel"`
	OccurredAt     time.Time `json:"occurred_at"`
}

func (e NotificationSent) EventMetadata() Metadata {
	return NewMetadata(EventNotificationSent, e.TenantID, e.CorrelationID, e.OccurredAt)
}
