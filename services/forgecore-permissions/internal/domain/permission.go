package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	ActionRead   = "read"
	ActionWrite  = "write"
	ActionDelete = "delete"
	ActionAdmin  = "admin"
)

type Role struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Permissions []string
	CreatedAt   time.Time
}

type Permission struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	UserID       uuid.UUID
	ResourceType string
	ResourceID   *uuid.UUID
	Action       string
	CreatedAt    time.Time
}

type RoleBinding struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	UserID    uuid.UUID
	RoleID    uuid.UUID
	CreatedAt time.Time
}
