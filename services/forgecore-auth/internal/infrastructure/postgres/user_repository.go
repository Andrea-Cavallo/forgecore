package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-auth/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
)

const userSelectCols = `id, tenant_id, email_enc, email_hash, password_hash, roles,
	mfa_enabled, mfa_secret, mfa_backup_codes, email_verified, locked_until,
	oauth_provider, oauth_provider_id, created_at, updated_at, deleted_at`

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	const q = `INSERT INTO users
		(id, tenant_id, email_enc, email_hash, password_hash, roles,
		 mfa_enabled, mfa_backup_codes, email_verified,
		 oauth_provider, oauth_provider_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`
	codes := u.MFABackupCodes
	if codes == nil {
		codes = []string{}
	}
	_, err := r.pool.Exec(ctx, q,
		u.ID, u.TenantID, u.EmailEnc, u.EmailHash, u.PasswordHash, u.Roles,
		u.MFAEnabled, codes, u.EmailVerified,
		u.OAuthProvider, u.OAuthProviderID, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.User, error) {
	q := `SELECT ` + userSelectCols + ` FROM users WHERE id=$1 AND tenant_id=$2 AND deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, q, id, tenantID)
	return scanUser(row)
}

func (r *UserRepository) GetByEmailHash(ctx context.Context, emailHash string, tenantID uuid.UUID) (*domain.User, error) {
	q := `SELECT ` + userSelectCols + ` FROM users WHERE email_hash=$1 AND tenant_id=$2 AND deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, q, emailHash, tenantID)
	return scanUser(row)
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	const q = `UPDATE users SET email_enc=$1, password_hash=$2, roles=$3,
		mfa_enabled=$4, mfa_secret=$5, mfa_backup_codes=$6,
		email_verified=$7, locked_until=$8,
		oauth_provider=$9, oauth_provider_id=$10, updated_at=$11
		WHERE id=$12 AND tenant_id=$13`
	codes := u.MFABackupCodes
	if codes == nil {
		codes = []string{}
	}
	_, err := r.pool.Exec(ctx, q,
		u.EmailEnc, u.PasswordHash, u.Roles,
		u.MFAEnabled, u.MFASecret, codes,
		u.EmailVerified, u.LockedUntil,
		u.OAuthProvider, u.OAuthProviderID, u.UpdatedAt,
		u.ID, u.TenantID)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	const q = `UPDATE users SET deleted_at=NOW() WHERE id=$1 AND tenant_id=$2`
	_, err := r.pool.Exec(ctx, q, id, tenantID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (r *UserRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*domain.User, error) {
	q := `SELECT ` + userSelectCols + ` FROM users WHERE tenant_id=$1 AND deleted_at IS NULL
		AND (created_at, id) < ($2, $3)
		ORDER BY created_at DESC, id DESC LIMIT 50`
	rows, err := r.pool.Query(ctx, q, tenantID, cursor.CreatedAt, cursor.ID)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	var users []*domain.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func scanUser(row pgx.Row) (*domain.User, error) {
	var u domain.User
	err := row.Scan(
		&u.ID, &u.TenantID, &u.EmailEnc, &u.EmailHash, &u.PasswordHash, &u.Roles,
		&u.MFAEnabled, &u.MFASecret, &u.MFABackupCodes, &u.EmailVerified, &u.LockedUntil,
		&u.OAuthProvider, &u.OAuthProviderID, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("scan user: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetByOAuthProvider(ctx context.Context, provider, providerID string, tenantID uuid.UUID) (*domain.User, error) {
	q := `SELECT ` + userSelectCols + ` FROM users
		WHERE oauth_provider=$1 AND oauth_provider_id=$2 AND tenant_id=$3 AND deleted_at IS NULL`
	row := r.pool.QueryRow(ctx, q, provider, providerID, tenantID)
	return scanUser(row)
}
