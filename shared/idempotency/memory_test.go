package idempotency

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryStoreReserveAndReplay(t *testing.T) {
	store := NewMemoryStore()
	record := Record{TenantID: "tenant", Operation: "payment", Key: "key", Fingerprint: "a"}

	_, reserved, err := store.Reserve(context.Background(), record)
	if err != nil || !reserved {
		t.Fatalf("prima reserve = %v %v", reserved, err)
	}
	_, reserved, err = store.Reserve(context.Background(), record)
	if err != nil || reserved {
		t.Fatalf("replay reserve = %v %v", reserved, err)
	}
}

func TestMemoryStoreDetectsConflict(t *testing.T) {
	store := NewMemoryStore()
	_, _, err := store.Reserve(context.Background(), Record{TenantID: "tenant", Operation: "payment", Key: "key", Fingerprint: "a"})
	if err != nil {
		t.Fatalf("reserve iniziale: %v", err)
	}
	_, _, err = store.Reserve(context.Background(), Record{TenantID: "tenant", Operation: "payment", Key: "key", Fingerprint: "b"})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("errore = %v", err)
	}
}
