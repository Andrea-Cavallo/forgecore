package postgres

import "testing"

func TestSetTenantSQL(t *testing.T) {
	if setTenantSQL != "SET LOCAL app.tenant_id = $1" {
		t.Fatalf("sql inatteso: %s", setTenantSQL)
	}
}
