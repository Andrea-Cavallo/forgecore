package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	UserID     uuid.UUID
	DeviceID   string
	UserAgent  string
	IPAddress  string
	LastSeenAt time.Time
	ExpiresAt  time.Time
}

func (s *Session) IsExpired(now time.Time) bool {
	return now.After(s.ExpiresAt)
}
