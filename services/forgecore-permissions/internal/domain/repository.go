package domain

import (
	"context"

	"github.com/google/uuid"
)

type PermissionRepository interface {
	Grant(ctx context.Context, p *Permission) error
	Revoke(ctx context.Context, id, tenantID uuid.UUID) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Permission, error)
	ListByUser(ctx context.Context, userID, tenantID uuid.UUID) ([]*Permission, error)
	CheckPermission(ctx context.Context, userID uuid.UUID, resourceType string, resourceID *uuid.UUID, action string, tenantID uuid.UUID) (bool, error)
}

type RoleRepository interface {
	Create(ctx context.Context, r *Role) error
	GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, name string, tenantID uuid.UUID) (*Role, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*Role, error)
	BindUser(ctx context.Context, rb *RoleBinding) error
	UnbindUser(ctx context.Context, userID, roleID, tenantID uuid.UUID) error
	ListUserRoles(ctx context.Context, userID, tenantID uuid.UUID) ([]*Role, error)
}
