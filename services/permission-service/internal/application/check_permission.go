package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/permission-service/internal/domain"
)

type CheckPermissionInput struct {
	TenantID     uuid.UUID
	UserID       uuid.UUID
	ResourceType string
	ResourceID   *uuid.UUID
	Action       string
}

func (i CheckPermissionInput) Validate() error {
	if i.ResourceType == "" {
		return fmt.Errorf("tipo risorsa obbligatorio")
	}
	if i.Action == "" {
		return fmt.Errorf("azione obbligatoria")
	}
	return nil
}

type CheckPermissionOutput struct {
	Allowed bool
	Reason  string
}

type CheckPermissionUseCase struct {
	perms domain.PermissionRepository
	roles domain.RoleRepository
}

func NewCheckPermissionUseCase(perms domain.PermissionRepository, roles domain.RoleRepository) *CheckPermissionUseCase {
	return &CheckPermissionUseCase{perms: perms, roles: roles}
}

func (uc *CheckPermissionUseCase) Execute(ctx context.Context, input CheckPermissionInput) (*CheckPermissionOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("input non valido: %w", err)
	}
	allowed, err := uc.perms.CheckPermission(ctx, input.UserID, input.ResourceType, input.ResourceID, input.Action, input.TenantID)
	if err != nil {
		return nil, fmt.Errorf("verifica permesso fallita: %w", err)
	}
	if allowed {
		return &CheckPermissionOutput{Allowed: true, Reason: "permesso diretto"}, nil
	}
	userRoles, err := uc.roles.ListUserRoles(ctx, input.UserID, input.TenantID)
	if err != nil {
		return nil, fmt.Errorf("lettura ruoli utente fallita: %w", err)
	}
	for _, role := range userRoles {
		for _, perm := range role.Permissions {
			if perm == input.Action || perm == domain.ActionAdmin {
				return &CheckPermissionOutput{Allowed: true, Reason: fmt.Sprintf("ruolo: %s", role.Name)}, nil
			}
		}
	}
	return &CheckPermissionOutput{Allowed: false, Reason: "nessun permesso trovato"}, nil
}
