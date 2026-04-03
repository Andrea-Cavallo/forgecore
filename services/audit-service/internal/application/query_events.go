package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/audit-service/internal/domain"
	"github.com/yourorg/golang-modules/shared/pagination"
)

type QueryEventsInput struct {
	TenantID uuid.UUID
	ActorID  *uuid.UUID
	Action   string
	Cursor   pagination.Cursor
	Limit    int
}

type QueryEventsUseCase struct {
	repo domain.AuditRepository
}

func NewQueryEventsUseCase(repo domain.AuditRepository) *QueryEventsUseCase {
	return &QueryEventsUseCase{repo: repo}
}

func (uc *QueryEventsUseCase) Execute(ctx context.Context, input QueryEventsInput) ([]*domain.AuditEntry, error) {
	if input.ActorID != nil {
		entries, err := uc.repo.ListByActor(ctx, *input.ActorID, input.TenantID, input.Cursor)
		if err != nil {
			return nil, fmt.Errorf("query eventi per attore fallita: %w", err)
		}
		return entries, nil
	}
	if input.Action != "" {
		entries, err := uc.repo.ListByAction(ctx, input.Action, input.TenantID, input.Cursor)
		if err != nil {
			return nil, fmt.Errorf("query eventi per azione fallita: %w", err)
		}
		return entries, nil
	}
	entries, err := uc.repo.ListByTenant(ctx, input.TenantID, input.Cursor)
	if err != nil {
		return nil, fmt.Errorf("query eventi per tenant fallita: %w", err)
	}
	return entries, nil
}
