package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/config-service/internal/domain"
)

type SetConfigInput struct {
	TenantID uuid.UUID
	Key      string
	Value    string
}

func (i SetConfigInput) Validate() error {
	if i.Key == "" {
		return fmt.Errorf("chiave obbligatoria")
	}
	return nil
}

type SetConfigUseCase struct {
	repo  domain.ConfigRepository
	cache domain.ConfigCache
}

func NewSetConfigUseCase(repo domain.ConfigRepository, cache domain.ConfigCache) *SetConfigUseCase {
	return &SetConfigUseCase{repo: repo, cache: cache}
}

func (uc *SetConfigUseCase) Execute(ctx context.Context, input SetConfigInput) error {
	if err := input.Validate(); err != nil {
		return fmt.Errorf("input non valido: %w", err)
	}
	cfg := &domain.TenantConfig{
		ID:        uuid.New(),
		TenantID:  input.TenantID,
		Key:       input.Key,
		Value:     input.Value,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := uc.repo.Set(ctx, cfg); err != nil {
		return fmt.Errorf("salvataggio configurazione fallito: %w", err)
	}
	_ = uc.cache.Invalidate(ctx, input.TenantID, input.Key)
	return nil
}
