package middleware

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestTenantContext(t *testing.T) {
	tenantID := uuid.New()
	ctx := WithTenant(context.Background(), tenantID)
	got, ok := TenantFromContext(ctx)
	if !ok || got != tenantID {
		t.Fatalf("tenant inatteso: %s", got)
	}
}
