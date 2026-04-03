package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

type TokenCleaner interface {
	CleanExpiredTokens(ctx context.Context, tenantID string) (int64, error)
}

type CleanupTokensHandler struct {
	cleaner TokenCleaner
}

func NewCleanupTokensHandler(cleaner TokenCleaner) *CleanupTokensHandler {
	return &CleanupTokensHandler{cleaner: cleaner}
}

func (h *CleanupTokensHandler) Handle(ctx context.Context, payload []byte) error {
	var p CleanupTokensPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("payload non valido: %w", err)
	}
	deleted, err := h.cleaner.CleanExpiredTokens(ctx, p.TenantID)
	if err != nil {
		return fmt.Errorf("pulizia token fallita: %w", err)
	}
	slog.Info("token scaduti rimossi", "tenant_id", p.TenantID, "count", deleted)
	return nil
}
