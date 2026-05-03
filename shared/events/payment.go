package events

import (
	"time"

	"github.com/google/uuid"
)

const (
	SubjectPaymentSucceeded = "payment.succeeded"
	SubjectPaymentFailed    = "payment.failed"
	SubjectPaymentRefunded  = "payment.refunded"
	SubjectInvoiceCreated   = "invoice.created"

	EventPaymentSucceeded = "payment.succeeded.v1"
	EventPaymentFailed    = "payment.failed.v1"
	EventPaymentRefunded  = "payment.refunded.v1"
	EventInvoiceCreated   = "invoice.created.v1"
)

// PaymentSucceeded is published when a payment is successfully processed.
type PaymentSucceeded struct {
	Version       int       `json:"version,omitempty"`
	EventName     string    `json:"event_name,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	TenantID      uuid.UUID `json:"tenant_id"`
	UserID        uuid.UUID `json:"user_id"`
	PaymentID     uuid.UUID `json:"payment_id"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Provider      string    `json:"provider"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e PaymentSucceeded) EventMetadata() Metadata {
	return NewMetadata(EventPaymentSucceeded, e.TenantID, e.CorrelationID, e.OccurredAt)
}

// PaymentFailed is published when a payment attempt fails.
type PaymentFailed struct {
	Version       int       `json:"version,omitempty"`
	EventName     string    `json:"event_name,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	TenantID      uuid.UUID `json:"tenant_id"`
	UserID        uuid.UUID `json:"user_id"`
	PaymentID     uuid.UUID `json:"payment_id"`
	Reason        string    `json:"reason"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e PaymentFailed) EventMetadata() Metadata {
	return NewMetadata(EventPaymentFailed, e.TenantID, e.CorrelationID, e.OccurredAt)
}

// PaymentRefunded is published when a refund is issued.
type PaymentRefunded struct {
	Version       int       `json:"version,omitempty"`
	EventName     string    `json:"event_name,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	TenantID      uuid.UUID `json:"tenant_id"`
	UserID        uuid.UUID `json:"user_id"`
	PaymentID     uuid.UUID `json:"payment_id"`
	Amount        int64     `json:"amount"`
	OccurredAt    time.Time `json:"occurred_at"`
}

func (e PaymentRefunded) EventMetadata() Metadata {
	return NewMetadata(EventPaymentRefunded, e.TenantID, e.CorrelationID, e.OccurredAt)
}
