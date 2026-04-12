package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/golang-modules/services/subscription-service/internal/domain"
)

type PlanRepository struct {
	pool *pgxpool.Pool
}

func NewPlanRepository(pool *pgxpool.Pool) *PlanRepository {
	return &PlanRepository{pool: pool}
}

func (r *PlanRepository) Create(ctx context.Context, p *domain.Plan) error {
	const q = `INSERT INTO plans (id, tenant_id, name, amount, currency, interval, provider_id, active, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, q, p.ID, p.TenantID, p.Name, p.Amount, p.Currency, p.Interval, p.ProviderID, p.Active, p.CreatedAt)
	if err != nil {
		return fmt.Errorf("create plan: %w", err)
	}
	return nil
}

func (r *PlanRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Plan, error) {
	const q = `SELECT id, tenant_id, name, amount, currency, interval, provider_id, active, created_at
		FROM plans WHERE id=$1 AND tenant_id=$2`
	row := r.pool.QueryRow(ctx, q, id, tenantID)
	return scanPlan(row)
}

func (r *PlanRepository) ListActive(ctx context.Context, tenantID uuid.UUID) ([]*domain.Plan, error) {
	const q = `SELECT id, tenant_id, name, amount, currency, interval, provider_id, active, created_at
		FROM plans WHERE tenant_id=$1 AND active=true ORDER BY amount ASC`
	rows, err := r.pool.Query(ctx, q, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list plans: %w", err)
	}
	defer rows.Close()
	var plans []*domain.Plan
	for rows.Next() {
		p, err := scanPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func scanPlan(row pgx.Row) (*domain.Plan, error) {
	var p domain.Plan
	err := row.Scan(&p.ID, &p.TenantID, &p.Name, &p.Amount, &p.Currency, &p.Interval, &p.ProviderID, &p.Active, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("piano non trovato")
		}
		return nil, fmt.Errorf("scan plan: %w", err)
	}
	return &p, nil
}
