package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type UserInfo struct {
	UserID    uuid.UUID
	TenantID  uuid.UUID
	Email     string
	Roles     []string
	Active    bool
	CreatedAt string
}

type UserClient interface {
	GetUser(ctx context.Context, userID, tenantID uuid.UUID) (*UserInfo, error)
	ListUsers(ctx context.Context, tenantID uuid.UUID, limit int) ([]*UserInfo, error)
	DisableUser(ctx context.Context, userID, tenantID uuid.UUID) error
}

type ManageUsersUseCase struct {
	client UserClient
}

func NewManageUsersUseCase(client UserClient) *ManageUsersUseCase {
	return &ManageUsersUseCase{client: client}
}

func (uc *ManageUsersUseCase) GetUser(ctx context.Context, userID, tenantID uuid.UUID) (*UserInfo, error) {
	user, err := uc.client.GetUser(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("lettura utente fallita: %w", err)
	}
	return user, nil
}

func (uc *ManageUsersUseCase) DisableUser(ctx context.Context, userID, tenantID uuid.UUID) error {
	if err := uc.client.DisableUser(ctx, userID, tenantID); err != nil {
		return fmt.Errorf("disabilitazione utente fallita: %w", err)
	}
	return nil
}
