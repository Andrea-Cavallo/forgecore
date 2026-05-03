package application

import (
	"context"
	"fmt"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-auth/internal/domain"
	"github.com/google/uuid"
)

type MeOutput struct {
	UserID        uuid.UUID
	TenantID      uuid.UUID
	Roles         []string
	EmailVerified bool
}

type MeUseCase struct {
	users domain.UserRepository
}

func NewMeUseCase(users domain.UserRepository) *MeUseCase {
	return &MeUseCase{users: users}
}

func (uc *MeUseCase) Execute(ctx context.Context, tenantID, userID uuid.UUID) (*MeOutput, error) {
	user, err := uc.users.GetByID(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("lettura utente corrente: %w", err)
	}
	return &MeOutput{
		UserID:        user.ID,
		TenantID:      user.TenantID,
		Roles:         user.Roles,
		EmailVerified: user.EmailVerified,
	}, nil
}
