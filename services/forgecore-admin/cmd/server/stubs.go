package main

import (
	"context"
	"errors"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-admin/internal/application"
	"github.com/google/uuid"
)

// stubTenantClient makes unavailable tenant dependencies explicit in local admin runs.
type stubTenantClient struct{}

func (s *stubTenantClient) GetTenantSummary(_ context.Context, _ uuid.UUID) (*application.TenantSummary, error) {
	return nil, errors.New("dipendenza tenant non configurata")
}

func (s *stubTenantClient) ListTenants(_ context.Context, _ int) ([]*application.TenantSummary, error) {
	return nil, errors.New("dipendenza tenant non configurata")
}

// stubUserClient makes unavailable user dependencies explicit in local admin runs.
type stubUserClient struct{}

func (s *stubUserClient) GetUser(_ context.Context, _, _ uuid.UUID) (*application.UserInfo, error) {
	return nil, errors.New("dipendenza utenti non configurata")
}

func (s *stubUserClient) ListUsers(_ context.Context, _ uuid.UUID, _ int) ([]*application.UserInfo, error) {
	return nil, errors.New("dipendenza utenti non configurata")
}

func (s *stubUserClient) DisableUser(_ context.Context, _, _ uuid.UUID) error {
	return errors.New("dipendenza utenti non configurata")
}

// stubStatsClient makes unavailable stats dependencies explicit in local admin runs.
type stubStatsClient struct{}

func (s *stubStatsClient) GetGlobalStats(_ context.Context) (*application.GlobalStats, error) {
	return nil, errors.New("dipendenza statistiche non configurata")
}

func (s *stubStatsClient) GetTenantStats(_ context.Context, _ uuid.UUID) (*application.TenantSummary, error) {
	return nil, errors.New("dipendenza statistiche non configurata")
}
