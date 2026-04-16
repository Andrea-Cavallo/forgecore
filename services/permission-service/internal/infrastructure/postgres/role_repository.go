package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Andrea-Cavallo/golang-modules/services/permission-service/internal/domain"
)

type RoleRepository struct {
	pool *pgxpool.Pool
}

func NewRoleRepository(pool *pgxpool.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	const q = `INSERT INTO roles (id, tenant_id, name, permissions, created_at)
		VALUES ($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, q, role.ID, role.TenantID, role.Name, role.Permissions, role.CreatedAt)
	if err != nil {
		return fmt.Errorf("create role: %w", err)
	}
	return nil
}

func (r *RoleRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Role, error) {
	const q = `SELECT id, tenant_id, name, permissions, created_at FROM roles WHERE id=$1 AND tenant_id=$2`
	row := r.pool.QueryRow(ctx, q, id, tenantID)
	return scanRole(row)
}

func (r *RoleRepository) GetByName(ctx context.Context, name string, tenantID uuid.UUID) (*domain.Role, error) {
	const q = `SELECT id, tenant_id, name, permissions, created_at FROM roles WHERE name=$1 AND tenant_id=$2`
	row := r.pool.QueryRow(ctx, q, name, tenantID)
	return scanRole(row)
}

func (r *RoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	const q = `SELECT id, tenant_id, name, permissions, created_at FROM roles WHERE tenant_id=$1 ORDER BY name`
	rows, err := r.pool.Query(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()
	var roles []*domain.Role
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *RoleRepository) BindUser(ctx context.Context, rb *domain.RoleBinding) error {
	const q = `INSERT INTO role_bindings (id, tenant_id, user_id, role_id, created_at)
		VALUES ($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, q, rb.ID, rb.TenantID, rb.UserID, rb.RoleID, rb.CreatedAt)
	if err != nil {
		return fmt.Errorf("bind user to role: %w", err)
	}
	return nil
}

func (r *RoleRepository) UnbindUser(ctx context.Context, userID, roleID, tenantID uuid.UUID) error {
	const q = `DELETE FROM role_bindings WHERE user_id=$1 AND role_id=$2 AND tenant_id=$3`
	_, err := r.pool.Exec(ctx, q, userID, roleID, tenantID)
	if err != nil {
		return fmt.Errorf("unbind user from role: %w", err)
	}
	return nil
}

func (r *RoleRepository) ListUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*domain.Role, error) {
	const q = `SELECT ro.id, ro.tenant_id, ro.name, ro.permissions, ro.created_at
		FROM roles ro
		JOIN role_bindings rb ON rb.role_id = ro.id
		WHERE rb.user_id=$1 AND rb.tenant_id=$2`
	rows, err := r.pool.Query(ctx, q, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list user roles: %w", err)
	}
	defer rows.Close()
	var roles []*domain.Role
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func scanRole(row pgx.Row) (*domain.Role, error) {
	var role domain.Role
	err := row.Scan(&role.ID, &role.TenantID, &role.Name, &role.Permissions, &role.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("ruolo non trovato")
		}
		return nil, fmt.Errorf("scan role: %w", err)
	}
	return &role, nil
}
