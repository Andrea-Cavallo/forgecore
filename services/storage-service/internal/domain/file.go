package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	BucketAvatars  = "avatars"
	BucketReceipts = "receipts"
	BucketExports  = "exports"

	MaxFileSizeBytes int64 = 100 * 1024 * 1024 // 100 MB
)

type File struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	UserID      uuid.UUID
	Bucket      string
	Key         string
	Filename    string
	ContentType string
	Size        int64
	CreatedAt   time.Time
}

type PresignedURL struct {
	URL       string
	ExpiresAt time.Time
}
