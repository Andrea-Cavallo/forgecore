package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Andrea-Cavallo/golang-modules/services/config-service/internal/domain"
)

type ConfigRepository struct {
	pool *pgxpool.Pool
}

func NewConfigRepository(pool *pgxpool.Pool) *ConfigRepository {
	return &ConfigRepository{pool: pool}
}

func (r *ConfigRepository) Get(ctx context.Context, tenantID uuid.UUID, key string) (*domain.TenantConfig, error) {
	const q = `SELECT id, tenant_id, key, value, created_at, updated_at
		FROM tenant_configs WHERE tenant_id=$1 AND key=$2`
	row := r.pool.QueryRow(ctx, q, tenantID, key)
	var cfg domain.TenantConfig
	err := row.Scan(&cfg.ID, &cfg.TenantID, &cfg.Key, &cfg.Value, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrConfigNotFound
		}
		return nil, fmt.Errorf("get config: %w", err)
	}
	return &cfg, nil
}

func (r *ConfigRepository) Set(ctx context.Context, cfg *domain.TenantConfig) error {
	const q = `INSERT INTO tenant_configs (id, tenant_id, key, value, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (tenant_id, key) DO UPDATE SET value=EXCLUDED.value, updated_at=EXCLUDED.updated_at`
	_, err := r.pool.Exec(ctx, q, cfg.ID, cfg.TenantID, cfg.Key, cfg.Value, cfg.CreatedAt, cfg.UpdatedAt)
	if err != nil {
		return fmt.Errorf("set config: %w", err)
	}
	return nil
}

func (r *ConfigRepository) Delete(ctx context.Context, tenantID uuid.UUID, key string) error {
	const q = `DELETE FROM tenant_configs WHERE tenant_id=$1 AND key=$2`
	_, err := r.pool.Exec(ctx, q, tenantID, key)
	if err != nil {
		return fmt.Errorf("delete config: %w", err)
	}
	return nil
}

func (r *ConfigRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.TenantConfig, error) {
	const q = `SELECT id, tenant_id, key, value, created_at, updated_at
		FROM tenant_configs WHERE tenant_id=$1 ORDER BY key ASC`
	rows, err := r.pool.Query(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list configs: %w", err)
	}
	defer rows.Close()
	var cfgs []*domain.TenantConfig
	for rows.Next() {
		var cfg domain.TenantConfig
		if err := rows.Scan(&cfg.ID, &cfg.TenantID, &cfg.Key, &cfg.Value, &cfg.CreatedAt, &cfg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan config: %w", err)
		}
		cfgs = append(cfgs, &cfg)
	}
	return cfgs, rows.Err()
}
