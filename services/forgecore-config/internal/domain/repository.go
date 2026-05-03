package domain

import (
	"context"

	"github.com/google/uuid"
)

type ConfigRepository interface {
	Get(ctx context.Context, tenantID uuid.UUID, key string) (*TenantConfig, error)
	Set(ctx context.Context, cfg *TenantConfig) error
	Delete(ctx context.Context, tenantID uuid.UUID, key string) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*TenantConfig, error)
}

type ConfigCache interface {
	Get(ctx context.Context, tenantID uuid.UUID, key string) (string, error)
	Set(ctx context.Context, tenantID uuid.UUID, key, value string, ttlSeconds int64) error
	Invalidate(ctx context.Context, tenantID uuid.UUID, key string) error
}
