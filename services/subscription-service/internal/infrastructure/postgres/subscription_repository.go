package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Andrea-Cavallo/golang-modules/services/subscription-service/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
)

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, s *domain.Subscription) error {
	const q = `INSERT INTO subscriptions
		(id, tenant_id, user_id, plan_id, status, provider_id, current_period_start, current_period_end, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.pool.Exec(ctx, q,
		s.ID, s.TenantID, s.UserID, s.PlanID, s.Status, s.ProviderID,
		s.CurrentPeriodStart, s.CurrentPeriodEnd, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Subscription, error) {
	const q = `SELECT id, tenant_id, user_id, plan_id, status, provider_id,
		current_period_start, current_period_end, canceled_at, created_at, updated_at
		FROM subscriptions WHERE id=$1 AND tenant_id=$2`
	row := r.pool.QueryRow(ctx, q, id, tenantID)
	return scanSubscription(row)
}

func (r *SubscriptionRepository) GetActiveByUser(ctx context.Context, userID, tenantID uuid.UUID) (*domain.Subscription, error) {
	const q = `SELECT id, tenant_id, user_id, plan_id, status, provider_id,
		current_period_start, current_period_end, canceled_at, created_at, updated_at
		FROM subscriptions WHERE user_id=$1 AND tenant_id=$2 AND status IN ('active','trialing') LIMIT 1`
	row := r.pool.QueryRow(ctx, q, userID, tenantID)
	return scanSubscription(row)
}

func (r *SubscriptionRepository) Update(ctx context.Context, s *domain.Subscription) error {
	const q = `UPDATE subscriptions SET status=$1, provider_id=$2, current_period_end=$3, canceled_at=$4, updated_at=$5
		WHERE id=$6 AND tenant_id=$7`
	_, err := r.pool.Exec(ctx, q,
		s.Status, s.ProviderID, s.CurrentPeriodEnd, s.CanceledAt, s.UpdatedAt, s.ID, s.TenantID)
	if err != nil {
		return fmt.Errorf("update subscription: %w", err)
	}
	return nil
}

func (r *SubscriptionRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*domain.Subscription, error) {
	const q = `SELECT id, tenant_id, user_id, plan_id, status, provider_id,
		current_period_start, current_period_end, canceled_at, created_at, updated_at
		FROM subscriptions WHERE tenant_id=$1 AND (created_at, id) < ($2, $3)
		ORDER BY created_at DESC, id DESC LIMIT 50`
	return r.querySubscriptions(ctx, q, tenantID, cursor.CreatedAt, cursor.ID)
}

func (r *SubscriptionRepository) ListExpiring(ctx context.Context, before time.Time) ([]*domain.Subscription, error) {
	const q = `SELECT id, tenant_id, user_id, plan_id, status, provider_id,
		current_period_start, current_period_end, canceled_at, created_at, updated_at
		FROM subscriptions WHERE status='active' AND current_period_end < $1 ORDER BY current_period_end ASC LIMIT 100`
	return r.querySubscriptions(ctx, q, before)
}

func (r *SubscriptionRepository) querySubscriptions(ctx context.Context, q string, args ...any) ([]*domain.Subscription, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query subscriptions: %w", err)
	}
	defer rows.Close()
	var subs []*domain.Subscription
	for rows.Next() {
		s, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

func scanSubscription(row pgx.Row) (*domain.Subscription, error) {
	var s domain.Subscription
	err := row.Scan(&s.ID, &s.TenantID, &s.UserID, &s.PlanID, &s.Status, &s.ProviderID,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CanceledAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("scan subscription: %w", err)
	}
	return &s, nil
}
