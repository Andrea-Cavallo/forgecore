package application

import (
	"context"
	"testing"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-permissions/internal/domain"
	"github.com/google/uuid"
)

type revokePermissionRepo struct {
	revokedID       uuid.UUID
	revokedTenantID uuid.UUID
}

func (r *revokePermissionRepo) Grant(context.Context, *domain.Permission) error {
	return nil
}

func (r *revokePermissionRepo) Revoke(_ context.Context, id, tenantID uuid.UUID) error {
	r.revokedID = id
	r.revokedTenantID = tenantID
	return nil
}

func (r *revokePermissionRepo) GetByID(context.Context, uuid.UUID, uuid.UUID) (*domain.Permission, error) {
	return nil, domain.ErrPermissionNotFound
}

func (r *revokePermissionRepo) ListByUser(context.Context, uuid.UUID, uuid.UUID) ([]*domain.Permission, error) {
	return nil, nil
}

func (r *revokePermissionRepo) CheckPermission(context.Context, uuid.UUID, string, *uuid.UUID, string, uuid.UUID) (bool, error) {
	return false, nil
}

func TestRevokePermissionUseCaseExecute(t *testing.T) {
	repo := &revokePermissionRepo{}
	uc := NewRevokePermissionUseCase(repo)
	tenantID := uuid.New()
	permissionID := uuid.New()

	err := uc.Execute(context.Background(), RevokePermissionInput{
		TenantID:     tenantID,
		PermissionID: permissionID,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if repo.revokedID != permissionID {
		t.Fatalf("revoked permission mismatch: got %s want %s", repo.revokedID, permissionID)
	}
	if repo.revokedTenantID != tenantID {
		t.Fatalf("revoked tenant mismatch: got %s want %s", repo.revokedTenantID, tenantID)
	}
}

func TestRevokePermissionUseCaseRejectsMissingIDs(t *testing.T) {
	repo := &revokePermissionRepo{}
	uc := NewRevokePermissionUseCase(repo)

	err := uc.Execute(context.Background(), RevokePermissionInput{})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
