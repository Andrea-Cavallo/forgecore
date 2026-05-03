package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-storage/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
)

type FileRepository struct {
	pool *pgxpool.Pool
}

func NewFileRepository(pool *pgxpool.Pool) *FileRepository {
	return &FileRepository{pool: pool}
}

func (r *FileRepository) Save(ctx context.Context, f *domain.File) error {
	const q = `INSERT INTO files (id, tenant_id, user_id, filename, bucket, key, content_type, size, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.pool.Exec(ctx, q,
		f.ID, f.TenantID, f.UserID, f.Filename, f.Bucket, f.Key,
		f.ContentType, f.Size, f.CreatedAt)
	if err != nil {
		return fmt.Errorf("save file: %w", err)
	}
	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*domain.File, error) {
	const q = `SELECT id, tenant_id, user_id, filename, bucket, key, content_type, size, created_at
		FROM files WHERE id=$1 AND tenant_id=$2`
	row := r.pool.QueryRow(ctx, q, id, tenantID)
	return scanFile(row)
}

func (r *FileRepository) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	const q = `DELETE FROM files WHERE id=$1 AND tenant_id=$2`
	_, err := r.pool.Exec(ctx, q, id, tenantID)
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

func (r *FileRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*domain.File, error) {
	const q = `SELECT id, tenant_id, user_id, filename, bucket, key, content_type, size, created_at
		FROM files WHERE tenant_id=$1 AND (created_at, id) < ($2, $3)
		ORDER BY created_at DESC, id DESC LIMIT 50`
	return r.queryFiles(ctx, q, tenantID, cursor.CreatedAt, cursor.ID)
}

func (r *FileRepository) ListByUser(ctx context.Context, userID, tenantID uuid.UUID, cursor pagination.Cursor) ([]*domain.File, error) {
	const q = `SELECT id, tenant_id, user_id, filename, bucket, key, content_type, size, created_at
		FROM files WHERE user_id=$1 AND tenant_id=$2 AND (created_at, id) < ($3, $4)
		ORDER BY created_at DESC, id DESC LIMIT 50`
	return r.queryFiles(ctx, q, userID, tenantID, cursor.CreatedAt, cursor.ID)
}

func (r *FileRepository) queryFiles(ctx context.Context, q string, args ...any) ([]*domain.File, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query files: %w", err)
	}
	defer rows.Close()
	var files []*domain.File
	for rows.Next() {
		f, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func scanFile(row pgx.Row) (*domain.File, error) {
	var f domain.File
	err := row.Scan(&f.ID, &f.TenantID, &f.UserID, &f.Filename, &f.Bucket, &f.Key,
		&f.ContentType, &f.Size, &f.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFileNotFound
		}
		return nil, fmt.Errorf("scan file: %w", err)
	}
	return &f, nil
}
