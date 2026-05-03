package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	ChannelEmail = "email"
	ChannelSMS   = "sms"

	StatusQueued = "queued"
	StatusSent   = "sent"
	StatusFailed = "failed"
)

type Notification struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	UserID    uuid.UUID
	Channel   string
	Template  string
	Recipient string
	Vars      map[string]string
	Status    string
	Attempts  int
	SentAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (n *Notification) CanRetry(maxAttempts int) bool {
	return n.Status == StatusFailed && n.Attempts < maxAttempts
}
