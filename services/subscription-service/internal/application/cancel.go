package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/subscription-service/internal/domain"
)

type CancelInput struct {
	SubscriptionID uuid.UUID
	TenantID       uuid.UUID
}

type CancelUseCase struct {
	subs    domain.SubscriptionRepository
	billing domain.BillingProvider
}

func NewCancelUseCase(subs domain.SubscriptionRepository, billing domain.BillingProvider) *CancelUseCase {
	return &CancelUseCase{subs: subs, billing: billing}
}

func (uc *CancelUseCase) Execute(ctx context.Context, input CancelInput) error {
	sub, err := uc.subs.GetByID(ctx, input.SubscriptionID, input.TenantID)
	if err != nil {
		return fmt.Errorf("abbonamento non trovato: %w", err)
	}
	if !sub.IsCancelable() {
		return domain.ErrSubscriptionNotActive
	}
	if err := uc.billing.CancelSubscription(ctx, sub.ProviderID); err != nil {
		return fmt.Errorf("cancellazione provider fallita: %w", err)
	}
	now := time.Now().UTC()
	sub.Status = domain.StatusCanceled
	sub.CanceledAt = &now
	sub.UpdatedAt = now
	if err := uc.subs.Update(ctx, sub); err != nil {
		return fmt.Errorf("aggiornamento abbonamento fallito: %w", err)
	}
	return nil
}
