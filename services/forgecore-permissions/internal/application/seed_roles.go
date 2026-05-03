package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-permissions/internal/domain"
)

const (
	RoleOwner          = "owner"
	RoleBillingManager = "billing-manager"
	RoleAdmin          = "admin"
	RoleReadOnly       = "read-only"
	RoleUser           = "user"
)

var defaultRoles = []struct {
	name        string
	permissions []string
}{
	{
		name: RoleOwner,
		permissions: []string{
			domain.ActionRead, domain.ActionWrite, domain.ActionDelete, domain.ActionAdmin,
		},
	},
	{
		name: RoleAdmin,
		permissions: []string{
			domain.ActionRead, domain.ActionWrite, domain.ActionDelete,
		},
	},
	{
		name: RoleBillingManager,
		permissions: []string{
			domain.ActionRead, domain.ActionWrite,
		},
	},
	{
		name: RoleReadOnly,
		permissions: []string{
			domain.ActionRead,
		},
	},
	{
		name: RoleUser,
		permissions: []string{
			domain.ActionRead, domain.ActionWrite,
		},
	},
}

// SeedRolesUseCase crea i ruoli predefiniti per un nuovo tenant.
type SeedRolesUseCase struct {
	roles domain.RoleRepository
}

func NewSeedRolesUseCase(roles domain.RoleRepository) *SeedRolesUseCase {
	return &SeedRolesUseCase{roles: roles}
}

func (uc *SeedRolesUseCase) Execute(ctx context.Context, tenantID uuid.UUID) error {
	for _, def := range defaultRoles {
		if err := uc.seedRole(ctx, tenantID, def.name, def.permissions); err != nil {
			return err
		}
	}
	return nil
}

func (uc *SeedRolesUseCase) seedRole(ctx context.Context, tenantID uuid.UUID, name string, perms []string) error {
	existing, err := uc.roles.GetByName(ctx, name, tenantID)
	if err == nil && existing != nil {
		return nil
	}
	role := &domain.Role{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        name,
		Permissions: perms,
	}
	if err := uc.roles.Create(ctx, role); err != nil {
		return fmt.Errorf("creazione ruolo %q per tenant %s fallita: %w", name, tenantID, err)
	}
	return nil
}
