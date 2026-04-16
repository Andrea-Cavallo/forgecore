package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/admin-service/internal/application"
)

// stubTenantClient is a placeholder until gRPC clients are implemented in Phase 7.
type stubTenantClient struct{}

func (s *stubTenantClient) GetTenantSummary(_ context.Context, _ uuid.UUID) (*application.TenantSummary, error) {
	return nil, fmt.Errorf("non implementato")
}

func (s *stubTenantClient) ListTenants(_ context.Context, _ int) ([]*application.TenantSummary, error) {
	return nil, fmt.Errorf("non implementato")
}

// stubUserClient is a placeholder until gRPC clients are implemented in Phase 7.
type stubUserClient struct{}

func (s *stubUserClient) GetUser(_ context.Context, _, _ uuid.UUID) (*application.UserInfo, error) {
	return nil, fmt.Errorf("non implementato")
}

func (s *stubUserClient) ListUsers(_ context.Context, _ uuid.UUID, _ int) ([]*application.UserInfo, error) {
	return nil, fmt.Errorf("non implementato")
}

func (s *stubUserClient) DisableUser(_ context.Context, _, _ uuid.UUID) error {
	return fmt.Errorf("non implementato")
}

// stubStatsClient is a placeholder until gRPC clients are implemented in Phase 7.
type stubStatsClient struct{}

func (s *stubStatsClient) GetGlobalStats(_ context.Context) (*application.GlobalStats, error) {
	return nil, fmt.Errorf("non implementato")
}

func (s *stubStatsClient) GetTenantStats(_ context.Context, _ uuid.UUID) (*application.TenantSummary, error) {
	return nil, fmt.Errorf("non implementato")
}
