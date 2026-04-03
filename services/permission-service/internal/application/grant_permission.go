package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/permission-service/internal/domain"
)

type GrantPermissionInput struct {
	TenantID     uuid.UUID
	UserID       uuid.UUID
	ResourceType string
	ResourceID   *uuid.UUID
	Action       string
}

func (i GrantPermissionInput) Validate() error {
	if i.ResourceType == "" {
		return fmt.Errorf("tipo risorsa obbligatorio")
	}
	if i.Action == "" {
		return fmt.Errorf("azione obbligatoria")
	}
	return nil
}

type GrantPermissionUseCase struct {
	perms domain.PermissionRepository
}

func NewGrantPermissionUseCase(perms domain.PermissionRepository) *GrantPermissionUseCase {
	return &GrantPermissionUseCase{perms: perms}
}

func (uc *GrantPermissionUseCase) Execute(ctx context.Context, input GrantPermissionInput) error {
	if err := input.Validate(); err != nil {
		return fmt.Errorf("input non valido: %w", err)
	}
	p := &domain.Permission{
		ID:           uuid.New(),
		TenantID:     input.TenantID,
		UserID:       input.UserID,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		Action:       input.Action,
		CreatedAt:    time.Now().UTC(),
	}
	if err := uc.perms.Grant(ctx, p); err != nil {
		return fmt.Errorf("concessione permesso fallita: %w", err)
	}
	return nil
}
