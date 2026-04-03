package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	SubjectNotificationRequested = "notification.requested"
	SubjectNotificationSent      = "notification.sent"
	SubjectNotificationFailed    = "notification.failed"
)

// NotificationRequested is published when any service needs to send a notification.
type NotificationRequested struct {
	TenantID   uuid.UUID         `json:"tenant_id"`
	UserID     uuid.UUID         `json:"user_id"`
	Channel    string            `json:"channel"` // "email", "sms"
	Template   string            `json:"template"`
	Vars       map[string]string `json:"vars"`
	OccurredAt time.Time         `json:"occurred_at"`
}

// NotificationSent is published when a notification is successfully delivered.
type NotificationSent struct {
	TenantID       uuid.UUID `json:"tenant_id"`
	NotificationID uuid.UUID `json:"notification_id"`
	UserID         uuid.UUID `json:"user_id"`
	Channel        string    `json:"channel"`
	OccurredAt     time.Time `json:"occurred_at"`
}
