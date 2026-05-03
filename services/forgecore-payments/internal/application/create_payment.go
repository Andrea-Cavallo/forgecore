package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-payments/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/events"
)

// CreatePaymentInput holds the data required to initiate a new payment.
type CreatePaymentInput struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	Amount     int64
	Currency   string
	CustomerID string
}

// Validate checks that all required fields have acceptable values.
func (i CreatePaymentInput) Validate() error {
	if i.Amount <= 0 {
		return fmt.Errorf("importo deve essere positivo")
	}
	if i.Currency == "" {
		return fmt.Errorf("valuta obbligatoria")
	}
	return nil
}

// CreatePaymentOutput carries the newly created Payment back to the caller.
type CreatePaymentOutput struct {
	Payment *domain.Payment
}

// CreatePaymentUseCase orchestrates charging via the provider and persisting the result.
type CreatePaymentUseCase struct {
	repo      domain.PaymentRepository
	provider  domain.PaymentProvider
	publisher *events.Publisher
}

// NewCreatePaymentUseCase wires up the dependencies for the use case.
func NewCreatePaymentUseCase(repo domain.PaymentRepository, provider domain.PaymentProvider, pub *events.Publisher) *CreatePaymentUseCase {
	return &CreatePaymentUseCase{repo: repo, provider: provider, publisher: pub}
}

// Execute runs the full payment creation flow: validate → charge → persist → publish.
func (uc *CreatePaymentUseCase) Execute(ctx context.Context, input CreatePaymentInput) (*CreatePaymentOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("input non valido: %w", err)
	}
	payment := uc.buildPayment(input)
	providerID, chargeErr := uc.provider.Charge(ctx, input.Amount, input.Currency, input.CustomerID)
	if chargeErr != nil {
		return nil, uc.handleChargeFailure(ctx, payment, chargeErr)
	}
	payment.Status = domain.StatusSucceeded
	payment.ProviderID = providerID
	if err := uc.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("salvataggio pagamento fallito: %w", err)
	}
	if err := uc.publisher.Publish(ctx, events.SubjectPaymentSucceeded, events.PaymentSucceeded{
		TenantID:   payment.TenantID,
		UserID:     payment.UserID,
		PaymentID:  payment.ID,
		Amount:     payment.Amount,
		Currency:   payment.Currency,
		Provider:   payment.Provider,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		return nil, fmt.Errorf("pubblicazione evento pagamento fallita: %w", err)
	}
	return &CreatePaymentOutput{Payment: payment}, nil
}

func (uc *CreatePaymentUseCase) buildPayment(input CreatePaymentInput) *domain.Payment {
	now := time.Now().UTC()
	return &domain.Payment{
		ID:        uuid.New(),
		TenantID:  input.TenantID,
		UserID:    input.UserID,
		Amount:    input.Amount,
		Currency:  input.Currency,
		Provider:  domain.ProviderStripe,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (uc *CreatePaymentUseCase) handleChargeFailure(ctx context.Context, payment *domain.Payment, chargeErr error) error {
	payment.Status = domain.StatusFailed
	payment.FailureReason = chargeErr.Error()
	if err := uc.repo.Create(ctx, payment); err != nil {
		return fmt.Errorf("salvataggio fallimento pagamento: %w", err)
	}
	if err := uc.publisher.Publish(ctx, events.SubjectPaymentFailed, events.PaymentFailed{
		TenantID:   payment.TenantID,
		UserID:     payment.UserID,
		PaymentID:  payment.ID,
		Reason:     payment.FailureReason,
		OccurredAt: time.Now().UTC(),
	}); err != nil {
		return fmt.Errorf("pubblicazione evento fallimento pagamento: %w", err)
	}
	return fmt.Errorf("addebito fallito: %w", chargeErr)
}
