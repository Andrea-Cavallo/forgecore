package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-payments/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/events"
)

// RefundInput holds the data required to issue a refund.
type RefundInput struct {
	PaymentID uuid.UUID
	TenantID  uuid.UUID
	Amount    int64
}

// RefundUseCase orchestrates the refund flow against the provider and DB.
type RefundUseCase struct {
	repo      domain.PaymentRepository
	provider  domain.PaymentProvider
	publisher *events.Publisher
}

// NewRefundUseCase wires up the dependencies for the refund use case.
func NewRefundUseCase(repo domain.PaymentRepository, provider domain.PaymentProvider, pub *events.Publisher) *RefundUseCase {
	return &RefundUseCase{repo: repo, provider: provider, publisher: pub}
}

// Execute loads the payment, validates refundability, calls the provider, and persists the new status.
func (uc *RefundUseCase) Execute(ctx context.Context, input RefundInput) error {
	payment, err := uc.repo.GetByID(ctx, input.PaymentID, input.TenantID)
	if err != nil {
		return fmt.Errorf("pagamento non trovato: %w", err)
	}

	if !payment.IsRefundable() {
		return domain.ErrPaymentNotRefundable
	}

	if err := uc.provider.Refund(ctx, payment.ProviderID, input.Amount); err != nil {
		return fmt.Errorf("rimborso fallito: %w", err)
	}

	payment.Status = domain.StatusRefunded
	payment.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Update(ctx, payment); err != nil {
		return fmt.Errorf("aggiornamento pagamento fallito: %w", err)
	}

	_ = uc.publisher.Publish(ctx, events.SubjectPaymentRefunded, events.PaymentRefunded{
		TenantID:   input.TenantID,
		UserID:     payment.UserID,
		PaymentID:  payment.ID,
		Amount:     input.Amount,
		OccurredAt: time.Now().UTC(),
	})

	return nil
}
