package events

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestEventMetadataDefaultsVersion(t *testing.T) {
	tenantID := uuid.New()
	meta := NewMetadata(EventPaymentSucceeded, tenantID, "req-1", time.Time{})
	if meta.Version != EventVersionV1 {
		t.Fatalf("versione inattesa: %d", meta.Version)
	}
	if meta.TenantID != tenantID {
		t.Fatalf("tenant inatteso: %s", meta.TenantID)
	}
	if meta.OccurredAt.IsZero() {
		t.Fatal("occurred_at non valorizzato")
	}
}

func TestPaymentSucceededIsVersioned(t *testing.T) {
	tenantID := uuid.New()
	event := PaymentSucceeded{TenantID: tenantID, CorrelationID: "corr"}
	meta := event.EventMetadata()
	if meta.EventName != EventPaymentSucceeded {
		t.Fatalf("nome evento inatteso: %s", meta.EventName)
	}
}
