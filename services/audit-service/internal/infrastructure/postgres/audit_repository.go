package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Andrea-Cavallo/golang-modules/services/audit-service/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
)

type AuditRepository struct {
	pool *pgxpool.Pool
}

func NewAuditRepository(pool *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) Append(ctx context.Context, e *domain.AuditEntry) error {
	const q = `INSERT INTO audit_entries
		(id, tenant_id, actor_id, actor_type, action, resource_id, resource_type, ip_address, metadata, occurred_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.pool.Exec(ctx, q,
		e.ID, e.TenantID, e.ActorID, e.ActorType, e.Action,
		e.ResourceID, e.ResourceType, e.IPAddress, e.Metadata, e.OccurredAt)
	if err != nil {
		return fmt.Errorf("append audit entry: %w", err)
	}
	return nil
}

func (r *AuditRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.AuditEntry, error) {
	const q = `SELECT id, tenant_id, actor_id, actor_type, action, resource_id, resource_type, ip_address, metadata, occurred_at
		FROM audit_entries WHERE id=$1 AND tenant_id=$2`
	row := r.pool.QueryRow(ctx, q, id, tenantID)
	return scanAuditEntry(row)
}

func (r *AuditRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*domain.AuditEntry, error) {
	const q = `SELECT id, tenant_id, actor_id, actor_type, action, resource_id, resource_type, ip_address, metadata, occurred_at
		FROM audit_entries WHERE tenant_id=$1 AND (occurred_at, id) < ($2, $3)
		ORDER BY occurred_at DESC, id DESC LIMIT 50`
	return r.queryEntries(ctx, q, tenantID, cursor.CreatedAt, cursor.ID)
}

func (r *AuditRepository) ListByActor(ctx context.Context, actorID, tenantID uuid.UUID, cursor pagination.Cursor) ([]*domain.AuditEntry, error) {
	const q = `SELECT id, tenant_id, actor_id, actor_type, action, resource_id, resource_type, ip_address, metadata, occurred_at
		FROM audit_entries WHERE actor_id=$1 AND tenant_id=$2 AND (occurred_at, id) < ($3, $4)
		ORDER BY occurred_at DESC, id DESC LIMIT 50`
	return r.queryEntries(ctx, q, actorID, tenantID, cursor.CreatedAt, cursor.ID)
}

func (r *AuditRepository) ListByAction(ctx context.Context, action string, tenantID uuid.UUID, cursor pagination.Cursor) ([]*domain.AuditEntry, error) {
	const q = `SELECT id, tenant_id, actor_id, actor_type, action, resource_id, resource_type, ip_address, metadata, occurred_at
		FROM audit_entries WHERE action=$1 AND tenant_id=$2 AND (occurred_at, id) < ($3, $4)
		ORDER BY occurred_at DESC, id DESC LIMIT 50`
	return r.queryEntries(ctx, q, action, tenantID, cursor.CreatedAt, cursor.ID)
}

func (r *AuditRepository) queryEntries(ctx context.Context, q string, args ...any) ([]*domain.AuditEntry, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query audit entries: %w", err)
	}
	defer rows.Close()
	var entries []*domain.AuditEntry
	for rows.Next() {
		e, err := scanAuditEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func scanAuditEntry(row pgx.Row) (*domain.AuditEntry, error) {
	var e domain.AuditEntry
	err := row.Scan(&e.ID, &e.TenantID, &e.ActorID, &e.ActorType, &e.Action,
		&e.ResourceID, &e.ResourceType, &e.IPAddress, &e.Metadata, &e.OccurredAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("audit entry not found")
		}
		return nil, fmt.Errorf("scan audit entry: %w", err)
	}
	return &e, nil
}
