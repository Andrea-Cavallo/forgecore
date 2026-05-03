package idempotency

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrConflict = errors.New("chiave idempotenza gia' usata con fingerprint diverso")

type Status string

const (
	StatusStarted   Status = "started"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type Record struct {
	TenantID    string
	Operation   string
	Key         string
	Fingerprint string
	Status      Status
	Response    []byte
	ExpiresAt   time.Time
}

type Store interface {
	Reserve(ctx context.Context, record Record) (Record, bool, error)
	Complete(ctx context.Context, tenantID string, operation string, key string, response []byte) error
	Fail(ctx context.Context, tenantID string, operation string, key string, reason string) error
}

func Scope(tenantID string, operation string, key string) string {
	return fmt.Sprintf("%s:%s:%s", tenantID, operation, key)
}
