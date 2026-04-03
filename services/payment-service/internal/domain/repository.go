package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/shared/pagination"
)

// PaymentRepository defines persistence operations for Payment entities.
type PaymentRepository interface {
	Create(ctx context.Context, p *Payment) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Payment, error)
	Update(ctx context.Context, p *Payment) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*Payment, error)
	ListByUser(ctx context.Context, userID, tenantID uuid.UUID, cursor pagination.Cursor) ([]*Payment, error)
}

// PaymentProvider abstracts the external payment gateway (e.g. Stripe).
type PaymentProvider interface {
	Charge(ctx context.Context, amount int64, currency, customerID string) (providerID string, err error)
	Refund(ctx context.Context, providerID string, amount int64) error
}
