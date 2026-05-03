package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-webhooks/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
)

type EndpointRepository struct {
	pool *pgxpool.Pool
}

func NewEndpointRepository(pool *pgxpool.Pool) *EndpointRepository {
	return &EndpointRepository{pool: pool}
}

func (r *EndpointRepository) Create(ctx context.Context, e *domain.WebhookEndpoint) error {
	const q = `INSERT INTO webhook_endpoints (id, tenant_id, url, secret, events, active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.pool.Exec(ctx, q, e.ID, e.TenantID, e.URL, e.Secret, e.Events, e.Active, e.CreatedAt, e.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create endpoint: %w", err)
	}
	return nil
}

func (r *EndpointRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.WebhookEndpoint, error) {
	const q = `SELECT id, tenant_id, url, secret, events, active, created_at, updated_at
		FROM webhook_endpoints WHERE id=$1 AND tenant_id=$2`
	row := r.pool.QueryRow(ctx, q, id, tenantID)
	return scanEndpoint(row)
}

func (r *EndpointRepository) Update(ctx context.Context, e *domain.WebhookEndpoint) error {
	const q = `UPDATE webhook_endpoints SET url=$1, events=$2, active=$3, updated_at=$4
		WHERE id=$5 AND tenant_id=$6`
	_, err := r.pool.Exec(ctx, q, e.URL, e.Events, e.Active, e.UpdatedAt, e.ID, e.TenantID)
	if err != nil {
		return fmt.Errorf("update endpoint: %w", err)
	}
	return nil
}

func (r *EndpointRepository) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	const q = `DELETE FROM webhook_endpoints WHERE id=$1 AND tenant_id=$2`
	_, err := r.pool.Exec(ctx, q, id, tenantID)
	if err != nil {
		return fmt.Errorf("delete endpoint: %w", err)
	}
	return nil
}

func (r *EndpointRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*domain.WebhookEndpoint, error) {
	const q = `SELECT id, tenant_id, url, secret, events, active, created_at, updated_at
		FROM webhook_endpoints WHERE tenant_id=$1 AND (created_at, id) < ($2, $3)
		ORDER BY created_at DESC, id DESC LIMIT 50`
	return r.queryEndpoints(ctx, q, tenantID, cursor.CreatedAt, cursor.ID)
}

func (r *EndpointRepository) ListActiveByEvent(ctx context.Context, eventType string) ([]*domain.WebhookEndpoint, error) {
	const q = `SELECT id, tenant_id, url, secret, events, active, created_at, updated_at
		FROM webhook_endpoints WHERE active=true AND ($1 = ANY(events) OR '*' = ANY(events))`
	return r.queryEndpoints(ctx, q, eventType)
}

func (r *EndpointRepository) queryEndpoints(ctx context.Context, q string, args ...any) ([]*domain.WebhookEndpoint, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query endpoints: %w", err)
	}
	defer rows.Close()
	var endpoints []*domain.WebhookEndpoint
	for rows.Next() {
		e, err := scanEndpoint(rows)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, e)
	}
	return endpoints, rows.Err()
}

func scanEndpoint(row pgx.Row) (*domain.WebhookEndpoint, error) {
	var e domain.WebhookEndpoint
	err := row.Scan(&e.ID, &e.TenantID, &e.URL, &e.Secret, &e.Events, &e.Active, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("endpoint not found")
		}
		return nil, fmt.Errorf("scan endpoint: %w", err)
	}
	return &e, nil
}
