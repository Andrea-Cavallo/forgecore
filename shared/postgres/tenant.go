package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const setTenantSQL = "SET LOCAL app.tenant_id = $1"

func WithTenantTx(ctx context.Context, pool *pgxpool.Pool, tenantID uuid.UUID, fn func(pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("apertura transazione tenant: %w", err)
	}
	defer rollbackTenantTx(ctx, tx)
	if _, err := tx.Exec(ctx, setTenantSQL, tenantID); err != nil {
		return fmt.Errorf("impostazione tenant postgres: %w", err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transazione tenant: %w", err)
	}
	return nil
}

func rollbackTenantTx(ctx context.Context, tx pgx.Tx) {
	err := tx.Rollback(ctx)
	if err == nil || err == pgx.ErrTxClosed {
		return
	}
}
