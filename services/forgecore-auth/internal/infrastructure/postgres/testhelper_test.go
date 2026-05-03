//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "../../../../migrations")

	ctr, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("auth_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts(
			filepath.Join(migrationsDir, "000001_create_users.up.sql"),
		),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		t.Fatalf("avvio container postgres fallito: %v", err)
	}
	t.Cleanup(func() {
		if err := ctr.Terminate(ctx); err != nil {
			t.Logf("terminate container: %v", err)
		}
	})

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connessione stringa fallita: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("connessione pool fallita: %v", err)
	}
	t.Cleanup(pool.Close)

	// Necessario per RLS: crea ruolo app e disabilita RLS per il superuser di test
	_, err = pool.Exec(ctx, fmt.Sprintf(`
		ALTER TABLE users DISABLE ROW LEVEL SECURITY;
		ALTER TABLE sessions DISABLE ROW LEVEL SECURITY;
		SET app.tenant_id = '%s';
	`, "00000000-0000-0000-0000-000000000000"))
	if err != nil {
		t.Fatalf("setup RLS fallito: %v", err)
	}

	return pool
}
