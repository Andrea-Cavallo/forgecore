package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type GlobalStats struct {
	TotalUsers    int64
	TotalPayments int64
	TotalRevenue  int64
	ActiveTenants int64
}

type StatsClient interface {
	GetGlobalStats(ctx context.Context) (*GlobalStats, error)
	GetTenantStats(ctx context.Context, tenantID uuid.UUID) (*TenantSummary, error)
}

type GetStatsUseCase struct {
	stats StatsClient
}

func NewGetStatsUseCase(stats StatsClient) *GetStatsUseCase {
	return &GetStatsUseCase{stats: stats}
}

func (uc *GetStatsUseCase) Execute(ctx context.Context) (*GlobalStats, error) {
	result, err := uc.stats.GetGlobalStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("lettura statistiche fallita: %w", err)
	}
	return result, nil
}
