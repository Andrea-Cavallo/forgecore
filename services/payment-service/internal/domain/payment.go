package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending   = "pending"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
	StatusRefunded  = "refunded"
)

// Payment represents a financial transaction within a tenant.
type Payment struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	UserID        uuid.UUID
	Amount        int64
	Currency      string
	Status        string
	Provider      string
	ProviderID    string
	FailureReason string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// IsRefundable returns true only when the payment has been successfully charged.
func (p *Payment) IsRefundable() bool {
	return p.Status == StatusSucceeded
}
