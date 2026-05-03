//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-auth/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-auth/internal/infrastructure/postgres"
	"github.com/Andrea-Cavallo/golang-modules/shared/pagination"
)

func TestUserRepo_CreateAndGetByEmailHash(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := postgres.NewUserRepository(db)

	tenantID := uuid.New()
	user := &domain.User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EmailEnc:      []byte("encrypted@example.com"),
		EmailHash:     "hash_abc123",
		PasswordHash:  "$2a$12$testhash",
		Roles:         []string{"user"},
		MFAEnabled:    false,
		EmailVerified: false,
		CreatedAt:     time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt:     time.Now().UTC().Truncate(time.Microsecond),
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := repo.GetByEmailHash(ctx, user.EmailHash, tenantID)
	if err != nil {
		t.Fatalf("GetByEmailHash: %v", err)
	}
	if found.ID != user.ID {
		t.Errorf("ID mismatch: got %v, want %v", found.ID, user.ID)
	}
	if found.EmailHash != user.EmailHash {
		t.Errorf("EmailHash mismatch: got %v, want %v", found.EmailHash, user.EmailHash)
	}
}

func TestUserRepo_GetByID(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := postgres.NewUserRepository(db)

	tenantID := uuid.New()
	user := &domain.User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EmailEnc:      []byte("enc2@example.com"),
		EmailHash:     "hash_getbyid",
		PasswordHash:  "$2a$12$testhash2",
		Roles:         []string{"admin"},
		EmailVerified: true,
		CreatedAt:     time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt:     time.Now().UTC().Truncate(time.Microsecond),
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	found, err := repo.GetByID(ctx, user.ID, tenantID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if found.ID != user.ID {
		t.Errorf("ID mismatch: got %v, want %v", found.ID, user.ID)
	}
	if !found.EmailVerified {
		t.Error("atteso email_verified=true")
	}
}

func TestUserRepo_GetByEmailHash_NotFound(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := postgres.NewUserRepository(db)

	_, err := repo.GetByEmailHash(ctx, "hash_inesistente", uuid.New())
	if err == nil {
		t.Fatal("atteso errore ErrUserNotFound")
	}
	if err != domain.ErrUserNotFound {
		t.Errorf("atteso ErrUserNotFound, ricevuto: %v", err)
	}
}

func TestUserRepo_Update(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := postgres.NewUserRepository(db)

	tenantID := uuid.New()
	user := &domain.User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EmailEnc:      []byte("upd@example.com"),
		EmailHash:     "hash_update",
		PasswordHash:  "$2a$12$originale",
		Roles:         []string{"user"},
		EmailVerified: false,
		CreatedAt:     time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt:     time.Now().UTC().Truncate(time.Microsecond),
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	user.EmailVerified = true
	user.Roles = []string{"user", "admin"}
	user.UpdatedAt = time.Now().UTC().Truncate(time.Microsecond)
	if err := repo.Update(ctx, user); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, err := repo.GetByID(ctx, user.ID, tenantID)
	if err != nil {
		t.Fatalf("GetByID dopo Update: %v", err)
	}
	if !updated.EmailVerified {
		t.Error("atteso email_verified=true dopo update")
	}
	if len(updated.Roles) != 2 {
		t.Errorf("attesi 2 ruoli, ricevuti %d", len(updated.Roles))
	}
}

func TestUserRepo_Delete(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := postgres.NewUserRepository(db)

	tenantID := uuid.New()
	user := &domain.User{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EmailEnc:      []byte("del@example.com"),
		EmailHash:     "hash_delete",
		PasswordHash:  "$2a$12$testhash_del",
		Roles:         []string{"user"},
		CreatedAt:     time.Now().UTC().Truncate(time.Microsecond),
		UpdatedAt:     time.Now().UTC().Truncate(time.Microsecond),
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.Delete(ctx, user.ID, tenantID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.GetByID(ctx, user.ID, tenantID)
	if err != domain.ErrUserNotFound {
		t.Errorf("atteso ErrUserNotFound dopo Delete, ricevuto: %v", err)
	}
}

func TestUserRepo_ListByTenant(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	repo := postgres.NewUserRepository(db)

	tenantID := uuid.New()
	for i := range 3 {
		u := &domain.User{
			ID:           uuid.New(),
			TenantID:     tenantID,
			EmailEnc:     []byte(fmt.Sprintf("list%d@example.com", i)),
			EmailHash:    fmt.Sprintf("hash_list_%d", i),
			PasswordHash: "$2a$12$listhash",
			Roles:        []string{"user"},
			CreatedAt:    time.Now().UTC().Truncate(time.Microsecond),
			UpdatedAt:    time.Now().UTC().Truncate(time.Microsecond),
		}
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("Create #%d: %v", i, err)
		}
	}

	// cursore con timestamp futuro — recupera tutti i record
	cursor := pagination.Cursor{
		ID:        uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff"),
		CreatedAt: time.Now().Add(time.Hour).UTC(),
	}
	users, err := repo.ListByTenant(ctx, tenantID, cursor)
	if err != nil {
		t.Fatalf("ListByTenant: %v", err)
	}
	if len(users) != 3 {
		t.Errorf("attesi 3 utenti, ricevuti %d", len(users))
	}
}
