package idempotency

import (
	"context"
	"sync"
)

type MemoryStore struct {
	mu      sync.Mutex
	records map[string]Record
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{records: map[string]Record{}}
}

func (s *MemoryStore) Reserve(_ context.Context, record Record) (Record, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	scope := Scope(record.TenantID, record.Operation, record.Key)
	existing, ok := s.records[scope]
	if !ok {
		record.Status = StatusStarted
		s.records[scope] = record
		return record, true, nil
	}
	if existing.Fingerprint != record.Fingerprint {
		return existing, false, ErrConflict
	}
	return existing, false, nil
}

func (s *MemoryStore) Complete(_ context.Context, tenantID string, operation string, key string, response []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	scope := Scope(tenantID, operation, key)
	record := s.records[scope]
	record.Status = StatusCompleted
	record.Response = response
	s.records[scope] = record
	return nil
}

func (s *MemoryStore) Fail(_ context.Context, tenantID string, operation string, key string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	scope := Scope(tenantID, operation, key)
	record := s.records[scope]
	record.Status = StatusFailed
	record.Response = []byte(reason)
	s.records[scope] = record
	return nil
}
