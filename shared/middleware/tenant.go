package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type tenantKey struct{}

const headerTenantID = "X-Tenant-ID"

// TenantFromContext retrieves the tenant ID stored in the context.
func TenantFromContext(ctx context.Context) (uuid.UUID, bool) {
	v, ok := ctx.Value(tenantKey{}).(uuid.UUID)
	return v, ok
}

// WithTenant stores a tenant ID in a context.
func WithTenant(ctx context.Context, tenantID uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenantID)
}

// TenantMiddleware reads X-Tenant-ID from the request header, validates it,
// and stores it in the request context. Returns 400 if missing or invalid.
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get(headerTenantID)
		if raw == "" {
			http.Error(w, "tenant_id mancante", http.StatusBadRequest)
			return
		}
		tenantID, err := uuid.Parse(raw)
		if err != nil {
			http.Error(w, "tenant_id non valido", http.StatusBadRequest)
			return
		}
		ctx := WithTenant(r.Context(), tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
