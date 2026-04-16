package domain

import (
	"context"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Notification, error)
	Update(ctx context.Context, n *Notification) error
	ListByUser(ctx context.Context, userID, tenantID uuid.UUID, cursor pagination.Cursor) ([]*Notification, error)
	ListPendingRetries(ctx context.Context, maxAttempts int) ([]*Notification, error)
}

type EmailProvider interface {
	Send(ctx context.Context, to, template string, vars map[string]string) error
}

type SMSProvider interface {
	Send(ctx context.Context, to, template string, vars map[string]string) error
}
