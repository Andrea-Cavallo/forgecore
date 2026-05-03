package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-permissions/internal/domain"
)

type PermissionRepository struct {
	pool *pgxpool.Pool
}

func NewPermissionRepository(pool *pgxpool.Pool) *PermissionRepository {
	return &PermissionRepository{pool: pool}
}

func (r *PermissionRepository) Grant(ctx context.Context, p *domain.Permission) error {
	const q = `INSERT INTO permissions (id, tenant_id, user_id, resource_type, resource_id, action, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, user_id, resource_type, resource_id, action) DO NOTHING`
	_, err := r.pool.Exec(ctx, q, p.ID, p.TenantID, p.UserID, p.ResourceType, p.ResourceID, p.Action, p.CreatedAt)
	if err != nil {
		return fmt.Errorf("grant permission: %w", err)
	}
	return nil
}

func (r *PermissionRepository) Revoke(ctx context.Context, id, tenantID uuid.UUID) error {
	const q = `DELETE FROM permissions WHERE id=$1 AND tenant_id=$2`
	_, err := r.pool.Exec(ctx, q, id, tenantID)
	if err != nil {
		return fmt.Errorf("revoke permission: %w", err)
	}
	return nil
}

func (r *PermissionRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Permission, error) {
	const q = `SELECT id, tenant_id, user_id, resource_type, resource_id, action, created_at
		FROM permissions WHERE id=$1 AND tenant_id=$2`
	row := r.pool.QueryRow(ctx, q, id, tenantID)
	var p domain.Permission
	err := row.Scan(&p.ID, &p.TenantID, &p.UserID, &p.ResourceType, &p.ResourceID, &p.Action, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPermissionNotFound
		}
		return nil, fmt.Errorf("get permission: %w", err)
	}
	return &p, nil
}

func (r *PermissionRepository) ListByUser(ctx context.Context, userID, tenantID uuid.UUID) ([]*domain.Permission, error) {
	const q = `SELECT id, tenant_id, user_id, resource_type, resource_id, action, created_at
		FROM permissions WHERE user_id=$1 AND tenant_id=$2 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	defer rows.Close()
	var perms []*domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.TenantID, &p.UserID, &p.ResourceType, &p.ResourceID, &p.Action, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms = append(perms, &p)
	}
	return perms, rows.Err()
}

func (r *PermissionRepository) CheckPermission(ctx context.Context, userID uuid.UUID, resourceType string, resourceID *uuid.UUID, action string, tenantID uuid.UUID) (bool, error) {
	const q = `SELECT COUNT(1) FROM permissions
		WHERE user_id=$1 AND tenant_id=$2 AND resource_type=$3 AND action=$4
		AND (resource_id=$5 OR resource_id IS NULL)`
	var count int
	err := r.pool.QueryRow(ctx, q, userID, tenantID, resourceType, action, resourceID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check permission: %w", err)
	}
	return count > 0, nil
}
