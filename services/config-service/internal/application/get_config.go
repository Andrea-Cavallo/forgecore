package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/config-service/internal/domain"
)

type GetConfigInput struct {
	TenantID uuid.UUID
	Key      string
}

func (i GetConfigInput) Validate() error {
	if i.Key == "" {
		return fmt.Errorf("chiave obbligatoria")
	}
	return nil
}

type GetConfigOutput struct {
	Value string
	Found bool
}

type GetConfigUseCase struct {
	repo  domain.ConfigRepository
	cache domain.ConfigCache
}

func NewGetConfigUseCase(repo domain.ConfigRepository, cache domain.ConfigCache) *GetConfigUseCase {
	return &GetConfigUseCase{repo: repo, cache: cache}
}

func (uc *GetConfigUseCase) Execute(ctx context.Context, input GetConfigInput) (*GetConfigOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("input non valido: %w", err)
	}
	cached, err := uc.cache.Get(ctx, input.TenantID, input.Key)
	if err == nil {
		return &GetConfigOutput{Value: cached, Found: true}, nil
	}
	cfg, err := uc.repo.Get(ctx, input.TenantID, input.Key)
	if err != nil {
		if errors.Is(err, domain.ErrConfigNotFound) {
			return &GetConfigOutput{Found: false}, nil
		}
		return nil, fmt.Errorf("lettura configurazione fallita: %w", err)
	}
	_ = uc.cache.Set(ctx, input.TenantID, input.Key, cfg.Value, domain.CacheTTLSeconds)
	return &GetConfigOutput{Value: cfg.Value, Found: true}, nil
}
