package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/shared/pagination"
)

type AuditRepository interface {
	Append(ctx context.Context, e *AuditEntry) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*AuditEntry, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, cursor pagination.Cursor) ([]*AuditEntry, error)
	ListByActor(ctx context.Context, actorID, tenantID uuid.UUID, cursor pagination.Cursor) ([]*AuditEntry, error)
	ListByAction(ctx context.Context, action string, tenantID uuid.UUID, cursor pagination.Cursor) ([]*AuditEntry, error)
}
