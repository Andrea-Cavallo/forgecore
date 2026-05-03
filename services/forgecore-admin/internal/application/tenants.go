package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type TenantSummary struct {
	TenantID     uuid.UUID
	UserCount    int64
	PaymentTotal int64
	ActiveSince  string
}

type TenantClient interface {
	GetTenantSummary(ctx context.Context, tenantID uuid.UUID) (*TenantSummary, error)
	ListTenants(ctx context.Context, limit int) ([]*TenantSummary, error)
}

type ListTenantsUseCase struct {
	client TenantClient
}

func NewListTenantsUseCase(client TenantClient) *ListTenantsUseCase {
	return &ListTenantsUseCase{client: client}
}

func (uc *ListTenantsUseCase) Execute(ctx context.Context, limit int) ([]*TenantSummary, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	tenants, err := uc.client.ListTenants(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("lettura tenant fallita: %w", err)
	}
	return tenants, nil
}
