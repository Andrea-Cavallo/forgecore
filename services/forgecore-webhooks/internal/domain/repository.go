package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
)

type EndpointRepository interface {
	Create(ctx context.Context, e *WebhookEndpoint) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*WebhookEndpoint, error)
	Update(ctx context.Context, e *WebhookEndpoint) error
	Delete(ctx context.Context, id, tenantID uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*WebhookEndpoint, error)
	ListActiveByEvent(ctx context.Context, eventType string) ([]*WebhookEndpoint, error)
}

type DeliveryRepository interface {
	Create(ctx context.Context, d *WebhookDelivery) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*WebhookDelivery, error)
	Update(ctx context.Context, d *WebhookDelivery) error
	ListByEndpoint(ctx context.Context, endpointID, tenantID uuid.UUID, cursor pagination.Cursor) ([]*WebhookDelivery, error)
	ListPendingRetries(ctx context.Context) ([]*WebhookDelivery, error)
}
