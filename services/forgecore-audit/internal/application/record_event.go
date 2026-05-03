package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-audit/internal/domain"
)

type RecordEventInput struct {
	TenantID     uuid.UUID
	ActorID      *uuid.UUID
	ActorType    string
	Action       string
	ResourceID   *uuid.UUID
	ResourceType string
	IPAddress    string
	Metadata     map[string]any
	OccurredAt   time.Time
}

func (i RecordEventInput) Validate() error {
	if i.Action == "" {
		return fmt.Errorf("azione obbligatoria")
	}
	if i.ActorType == "" {
		return fmt.Errorf("tipo attore obbligatorio")
	}
	return nil
}

type RecordEventUseCase struct {
	repo domain.AuditRepository
}

func NewRecordEventUseCase(repo domain.AuditRepository) *RecordEventUseCase {
	return &RecordEventUseCase{repo: repo}
}

func (uc *RecordEventUseCase) Execute(ctx context.Context, input RecordEventInput) error {
	if err := input.Validate(); err != nil {
		return fmt.Errorf("input non valido: %w", err)
	}
	entry := &domain.AuditEntry{
		ID:           uuid.New(),
		TenantID:     input.TenantID,
		ActorID:      input.ActorID,
		ActorType:    input.ActorType,
		Action:       input.Action,
		ResourceID:   input.ResourceID,
		ResourceType: input.ResourceType,
		IPAddress:    input.IPAddress,
		Metadata:     input.Metadata,
		OccurredAt:   func() time.Time {
			if !input.OccurredAt.IsZero() {
				return input.OccurredAt
			}
			return time.Now().UTC()
		}(),
	}
	if err := uc.repo.Append(ctx, entry); err != nil {
		return fmt.Errorf("registrazione evento audit fallita: %w", err)
	}
	return nil
}
