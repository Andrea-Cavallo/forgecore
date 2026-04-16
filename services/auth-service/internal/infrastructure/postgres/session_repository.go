package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Andrea-Cavallo/golang-modules/services/auth-service/internal/domain"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	const q = `INSERT INTO sessions
		(id, tenant_id, user_id, device_id, user_agent, ip_address, last_seen_at, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.pool.Exec(ctx, q,
		s.ID, s.TenantID, s.UserID, s.DeviceID, s.UserAgent, s.IPAddress,
		s.LastSeenAt, s.ExpiresAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.Session, error) {
	const q = `SELECT id, tenant_id, user_id, device_id, user_agent, ip_address, last_seen_at, expires_at
		FROM sessions WHERE id=$1 AND tenant_id=$2`
	row := r.pool.QueryRow(ctx, q, id, tenantID)
	var s domain.Session
	if err := row.Scan(&s.ID, &s.TenantID, &s.UserID, &s.DeviceID, &s.UserAgent,
		&s.IPAddress, &s.LastSeenAt, &s.ExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &s, nil
}

func (r *SessionRepository) ListByUser(ctx context.Context, userID, tenantID uuid.UUID) ([]*domain.Session, error) {
	const q = `SELECT id, tenant_id, user_id, device_id, user_agent, ip_address, last_seen_at, expires_at
		FROM sessions WHERE user_id=$1 AND tenant_id=$2 AND expires_at > NOW()
		ORDER BY last_seen_at DESC`
	rows, err := r.pool.Query(ctx, q, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()
	var sessions []*domain.Session
	for rows.Next() {
		var s domain.Session
		if err := rows.Scan(&s.ID, &s.TenantID, &s.UserID, &s.DeviceID, &s.UserAgent,
			&s.IPAddress, &s.LastSeenAt, &s.ExpiresAt); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, &s)
	}
	return sessions, rows.Err()
}

func (r *SessionRepository) DeleteByID(ctx context.Context, id, tenantID uuid.UUID) error {
	const q = `DELETE FROM sessions WHERE id=$1 AND tenant_id=$2`
	_, err := r.pool.Exec(ctx, q, id, tenantID)
	if err != nil {
		return fmt.Errorf("delete session by id: %w", err)
	}
	return nil
}

func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID, tenantID uuid.UUID) error {
	const q = `DELETE FROM sessions WHERE user_id=$1 AND tenant_id=$2`
	_, err := r.pool.Exec(ctx, q, userID, tenantID)
	if err != nil {
		return fmt.Errorf("delete sessions: %w", err)
	}
	return nil
}

func (r *SessionRepository) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE sessions SET last_seen_at=$1 WHERE id=$2`
	_, err := r.pool.Exec(ctx, q, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("update last seen: %w", err)
	}
	return nil
}
