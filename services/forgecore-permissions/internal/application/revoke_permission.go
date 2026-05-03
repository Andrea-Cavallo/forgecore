package application

import (
	"context"
	"fmt"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-permissions/internal/domain"
	"github.com/google/uuid"
)

type RevokePermissionInput struct {
	TenantID     uuid.UUID
	PermissionID uuid.UUID
}

func (i RevokePermissionInput) Validate() error {
	if i.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id obbligatorio")
	}
	if i.PermissionID == uuid.Nil {
		return fmt.Errorf("permission_id obbligatorio")
	}
	return nil
}

type RevokePermissionUseCase struct {
	perms domain.PermissionRepository
}

func NewRevokePermissionUseCase(perms domain.PermissionRepository) *RevokePermissionUseCase {
	return &RevokePermissionUseCase{perms: perms}
}

func (uc *RevokePermissionUseCase) Execute(ctx context.Context, input RevokePermissionInput) error {
	if err := input.Validate(); err != nil {
		return fmt.Errorf("input non valido: %w", err)
	}
	if err := uc.perms.Revoke(ctx, input.PermissionID, input.TenantID); err != nil {
		return fmt.Errorf("revoca permesso fallita: %w", err)
	}
	return nil
}
