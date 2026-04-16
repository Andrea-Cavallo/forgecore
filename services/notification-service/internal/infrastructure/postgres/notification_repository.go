package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Andrea-Cavallo/golang-modules/services/notification-service/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
)

type NotificationRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationRepository(pool *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{pool: pool}
}

func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	const q = `INSERT INTO notifications
		(id, tenant_id, user_id, channel, template, recipient, vars, status, attempts, sent_at, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`
	_, err := r.pool.Exec(ctx, q,
		n.ID, n.TenantID, n.UserID, n.Channel, n.Template,
		n.Recipient, n.Vars, n.Status, n.Attempts, n.SentAt,
		n.CreatedAt, n.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create notification: %w", err)
	}
	return nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Notification, error) {
	const q = `SELECT id, tenant_id, user_id, channel, template, recipient, vars, status, attempts, sent_at, created_at, updated_at
		FROM notifications WHERE id=$1 AND tenant_id=$2`
	row := r.pool.QueryRow(ctx, q, id, tenantID)
	return scanNotification(row)
}

func (r *NotificationRepository) Update(ctx context.Context, n *domain.Notification) error {
	const q = `UPDATE notifications SET status=$1, attempts=$2, sent_at=$3, updated_at=$4
		WHERE id=$5 AND tenant_id=$6`
	_, err := r.pool.Exec(ctx, q, n.Status, n.Attempts, n.SentAt, n.UpdatedAt, n.ID, n.TenantID)
	if err != nil {
		return fmt.Errorf("update notification: %w", err)
	}
	return nil
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID, tenantID uuid.UUID, cursor pagination.Cursor) ([]*domain.Notification, error) {
	const q = `SELECT id, tenant_id, user_id, channel, template, recipient, vars, status, attempts, sent_at, created_at, updated_at
		FROM notifications WHERE user_id=$1 AND tenant_id=$2 AND (created_at, id) < ($3, $4)
		ORDER BY created_at DESC, id DESC LIMIT 50`
	rows, err := r.pool.Query(ctx, q, userID, tenantID, cursor.CreatedAt, cursor.ID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()
	var ns []*domain.Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		ns = append(ns, n)
	}
	return ns, rows.Err()
}

func (r *NotificationRepository) ListPendingRetries(ctx context.Context, maxAttempts int) ([]*domain.Notification, error) {
	const q = `SELECT id, tenant_id, user_id, channel, template, recipient, vars, status, attempts, sent_at, created_at, updated_at
		FROM notifications WHERE status='failed' AND attempts < $1 ORDER BY created_at ASC LIMIT 100`
	rows, err := r.pool.Query(ctx, q, maxAttempts)
	if err != nil {
		return nil, fmt.Errorf("list pending retries: %w", err)
	}
	defer rows.Close()
	var ns []*domain.Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		ns = append(ns, n)
	}
	return ns, rows.Err()
}

func scanNotification(row pgx.Row) (*domain.Notification, error) {
	var n domain.Notification
	err := row.Scan(&n.ID, &n.TenantID, &n.UserID, &n.Channel, &n.Template,
		&n.Recipient, &n.Vars, &n.Status, &n.Attempts, &n.SentAt,
		&n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotificationNotFound
		}
		return nil, fmt.Errorf("scan notification: %w", err)
	}
	return &n, nil
}
