package domain

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/shared/pagination"
)

type FileRepository interface {
	Save(ctx context.Context, f *File) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*File, error)
	Delete(ctx context.Context, id, tenantID uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*File, error)
	ListByUser(ctx context.Context, userID, tenantID uuid.UUID, cursor pagination.Cursor) ([]*File, error)
}

type StorageProvider interface {
	Upload(ctx context.Context, bucket, key string, r io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, bucket, key string) error
	Presign(ctx context.Context, bucket, key string, ttlSeconds int64) (*PresignedURL, error)
}
