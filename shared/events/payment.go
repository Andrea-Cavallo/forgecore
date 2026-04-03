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
)

// PaymentSucceeded is published when a payment is successfully processed.
type PaymentSucceeded struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	UserID     uuid.UUID `json:"user_id"`
	PaymentID  uuid.UUID `json:"payment_id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	Provider   string    `json:"provider"`
	OccurredAt time.Time `json:"occurred_at"`
}

// PaymentFailed is published when a payment attempt fails.
type PaymentFailed struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	UserID     uuid.UUID `json:"user_id"`
	PaymentID  uuid.UUID `json:"payment_id"`
	Reason     string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}

// PaymentRefunded is published when a refund is issued.
type PaymentRefunded struct {
	TenantID   uuid.UUID `json:"tenant_id"`
	UserID     uuid.UUID `json:"user_id"`
	PaymentID  uuid.UUID `json:"payment_id"`
	Amount     int64     `json:"amount"`
	OccurredAt time.Time `json:"occurred_at"`
}
