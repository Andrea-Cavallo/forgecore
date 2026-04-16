package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Andrea-Cavallo/golang-modules/services/payment-service/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	const q = `INSERT INTO payments
		(id, tenant_id, user_id, amount, currency, status, provider, provider_id, failure_reason, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`
	_, err := r.pool.Exec(ctx, q,
		p.ID, p.TenantID, p.UserID, p.Amount, p.Currency,
		p.Status, p.Provider, p.ProviderID, p.FailureReason,
		p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Payment, error) {
	const q = `SELECT id, tenant_id, user_id, amount, currency, status, provider, provider_id, failure_reason, created_at, updated_at
		FROM payments WHERE id=$1 AND tenant_id=$2`
	row := r.pool.QueryRow(ctx, q, id, tenantID)
	return scanPayment(row)
}

func (r *PaymentRepository) Update(ctx context.Context, p *domain.Payment) error {
	const q = `UPDATE payments SET status=$1, provider_id=$2, failure_reason=$3, updated_at=$4
		WHERE id=$5 AND tenant_id=$6`
	_, err := r.pool.Exec(ctx, q, p.Status, p.ProviderID, p.FailureReason, p.UpdatedAt, p.ID, p.TenantID)
	if err != nil {
		return fmt.Errorf("update payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) GetByProviderID(ctx context.Context, provider, providerID string) (*domain.Payment, error) {
	const q = `SELECT id, tenant_id, user_id, amount, currency, status, provider, provider_id, failure_reason, created_at, updated_at
		FROM payments WHERE provider=$1 AND provider_id=$2`
	row := r.pool.QueryRow(ctx, q, provider, providerID)
	return scanPayment(row)
}

func (r *PaymentRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*domain.Payment, error) {
	const q = `SELECT id, tenant_id, user_id, amount, currency, status, provider, provider_id, failure_reason, created_at, updated_at
		FROM payments WHERE tenant_id=$1 AND (created_at, id) < ($2, $3)
		ORDER BY created_at DESC, id DESC LIMIT 50`
	return r.queryPayments(ctx, q, tenantID, cursor.CreatedAt, cursor.ID)
}

func (r *PaymentRepository) ListByUser(ctx context.Context, userID, tenantID uuid.UUID, cursor pagination.Cursor) ([]*domain.Payment, error) {
	const q = `SELECT id, tenant_id, user_id, amount, currency, status, provider, provider_id, failure_reason, created_at, updated_at
		FROM payments WHERE user_id=$1 AND tenant_id=$2 AND (created_at, id) < ($3, $4)
		ORDER BY created_at DESC, id DESC LIMIT 50`
	return r.queryPayments(ctx, q, userID, tenantID, cursor.CreatedAt, cursor.ID)
}

func (r *PaymentRepository) queryPayments(ctx context.Context, q string, args ...any) ([]*domain.Payment, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query payments: %w", err)
	}
	defer rows.Close()
	var payments []*domain.Payment
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

func scanPayment(row pgx.Row) (*domain.Payment, error) {
	var p domain.Payment
	err := row.Scan(&p.ID, &p.TenantID, &p.UserID, &p.Amount, &p.Currency,
		&p.Status, &p.Provider, &p.ProviderID, &p.FailureReason, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, fmt.Errorf("scan payment: %w", err)
	}
	return &p, nil
}
