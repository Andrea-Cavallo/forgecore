package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/golang-modules/services/webhook-service/internal/domain"
	"github.com/yourorg/golang-modules/shared/pagination"
)

type DeliveryRepository struct {
	pool *pgxpool.Pool
}

func NewDeliveryRepository(pool *pgxpool.Pool) *DeliveryRepository {
	return &DeliveryRepository{pool: pool}
}

func (r *DeliveryRepository) Create(ctx context.Context, d *domain.WebhookDelivery) error {
	const q = `INSERT INTO webhook_deliveries
		(id, tenant_id, endpoint_id, event_type, payload, status, attempts, last_error, delivered_at, next_retry_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err := r.pool.Exec(ctx, q,
		d.ID, d.TenantID, d.EndpointID, d.EventType, d.Payload,
		d.Status, d.Attempts, d.LastError, d.DeliveredAt, d.NextRetryAt,
		d.CreatedAt, d.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create delivery: %w", err)
	}
	return nil
}

func (r *DeliveryRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.WebhookDelivery, error) {
	const q = `SELECT id, tenant_id, endpoint_id, event_type, payload, status, attempts, last_error, delivered_at, next_retry_at, created_at, updated_at
		FROM webhook_deliveries WHERE id=$1 AND tenant_id=$2`
	row := r.pool.QueryRow(ctx, q, id, tenantID)
	return scanDelivery(row)
}

func (r *DeliveryRepository) Update(ctx context.Context, d *domain.WebhookDelivery) error {
	const q = `UPDATE webhook_deliveries
		SET status=$1, attempts=$2, last_error=$3, delivered_at=$4, next_retry_at=$5, updated_at=$6
		WHERE id=$7 AND tenant_id=$8`
	_, err := r.pool.Exec(ctx, q,
		d.Status, d.Attempts, d.LastError, d.DeliveredAt, d.NextRetryAt, d.UpdatedAt,
		d.ID, d.TenantID)
	if err != nil {
		return fmt.Errorf("update delivery: %w", err)
	}
	return nil
}

func (r *DeliveryRepository) ListByEndpoint(ctx context.Context, endpointID, tenantID uuid.UUID, cursor pagination.Cursor) ([]*domain.WebhookDelivery, error) {
	const q = `SELECT id, tenant_id, endpoint_id, event_type, payload, status, attempts, last_error, delivered_at, next_retry_at, created_at, updated_at
		FROM webhook_deliveries WHERE endpoint_id=$1 AND tenant_id=$2 AND (created_at, id) < ($3, $4)
		ORDER BY created_at DESC, id DESC LIMIT 50`
	return r.queryDeliveries(ctx, q, endpointID, tenantID, cursor.CreatedAt, cursor.ID)
}

func (r *DeliveryRepository) ListPendingRetries(ctx context.Context) ([]*domain.WebhookDelivery, error) {
	const q = `SELECT id, tenant_id, endpoint_id, event_type, payload, status, attempts, last_error, delivered_at, next_retry_at, created_at, updated_at
		FROM webhook_deliveries WHERE status='failed' AND attempts < $1 AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY created_at ASC LIMIT 100`
	return r.queryDeliveries(ctx, q, domain.MaxDeliveryAttempts)
}

func (r *DeliveryRepository) queryDeliveries(ctx context.Context, q string, args ...any) ([]*domain.WebhookDelivery, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query deliveries: %w", err)
	}
	defer rows.Close()
	var deliveries []*domain.WebhookDelivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, rows.Err()
}

func scanDelivery(row pgx.Row) (*domain.WebhookDelivery, error) {
	var d domain.WebhookDelivery
	err := row.Scan(&d.ID, &d.TenantID, &d.EndpointID, &d.EventType, &d.Payload,
		&d.Status, &d.Attempts, &d.LastError, &d.DeliveredAt, &d.NextRetryAt,
		&d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("delivery non trovata")
		}
		return nil, fmt.Errorf("scan delivery: %w", err)
	}
	return &d, nil
}
