package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, s *Subscription) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Subscription, error)
	GetActiveByUser(ctx context.Context, userID, tenantID uuid.UUID) (*Subscription, error)
	Update(ctx context.Context, s *Subscription) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*Subscription, error)
	ListExpiring(ctx context.Context, before time.Time) ([]*Subscription, error)
}

type PlanRepository interface {
	Create(ctx context.Context, p *Plan) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Plan, error)
	ListActive(ctx context.Context, tenantID uuid.UUID) ([]*Plan, error)
}

type BillingProvider interface {
	CreateSubscription(ctx context.Context, customerID, planProviderID string) (providerID string, err error)
	CancelSubscription(ctx context.Context, providerID string) error
	ChangeSubscriptionPlan(ctx context.Context, providerID, newPlanID string) error
}
