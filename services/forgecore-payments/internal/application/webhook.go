package application

import (
	"context"
	"fmt"
	"time"

	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-payments/internal/domain"
	"github.com/Andrea-Cavallo/golang-modules/shared/events"
)

const (
	stripeEventPaymentSucceeded = "payment_intent.succeeded"
	stripeEventPaymentFailed    = "payment_intent.payment_failed"
)

// WebhookVerifier is implemented by the Stripe provider (and can be mocked in tests).
type WebhookVerifier interface {
	VerifyWebhook(rawBody []byte, sigHeader string) (*domain.PaymentWebhookEvent, error)
}

// HandleStripeWebhookUseCase processes inbound Stripe webhook events.
type HandleStripeWebhookUseCase struct {
	verifier  WebhookVerifier
	repo      domain.PaymentRepository
	publisher *events.Publisher
}

// NewHandleStripeWebhookUseCase constructs the use case.
func NewHandleStripeWebhookUseCase(v WebhookVerifier, repo domain.PaymentRepository, pub *events.Publisher) *HandleStripeWebhookUseCase {
	return &HandleStripeWebhookUseCase{verifier: v, repo: repo, publisher: pub}
}

// HandleWebhookInput carries the raw request body and Stripe-Signature header.
type HandleWebhookInput struct {
	RawBody   []byte
	SigHeader string
}

// Execute verifies the webhook signature, maps the event to a domain transition, and persists it.
func (uc *HandleStripeWebhookUseCase) Execute(ctx context.Context, input HandleWebhookInput) error {
	event, err := uc.verifier.VerifyWebhook(input.RawBody, input.SigHeader)
	if err != nil {
		return fmt.Errorf("verifica webhook fallita: %w", err)
	}

	payment, err := uc.repo.GetByProviderID(ctx, domain.ProviderStripe, event.ProviderPaymentID)
	if err != nil {
		return fmt.Errorf("pagamento non trovato per provider_id %s: %w", event.ProviderPaymentID, err)
	}

	switch event.Type {
	case stripeEventPaymentSucceeded:
		return uc.markSucceeded(ctx, payment)
	case stripeEventPaymentFailed:
		return uc.markFailed(ctx, payment)
	default:
		// Ignore unknown event types — return nil so Stripe does not retry.
		return nil
	}
}

func (uc *HandleStripeWebhookUseCase) markSucceeded(ctx context.Context, p *domain.Payment) error {
	if p.Status == domain.StatusSucceeded {
		return nil // idempotent
	}
	p.Status = domain.StatusSucceeded
	p.UpdatedAt = time.Now().UTC()
	if err := uc.repo.Update(ctx, p); err != nil {
		return fmt.Errorf("aggiornamento pagamento riuscito: %w", err)
	}
	_ = uc.publisher.Publish(ctx, events.SubjectPaymentSucceeded, events.PaymentSucceeded{
		TenantID:   p.TenantID,
		UserID:     p.UserID,
		PaymentID:  p.ID,
		Amount:     p.Amount,
		Currency:   p.Currency,
		Provider:   p.Provider,
		OccurredAt: p.UpdatedAt,
	})
	return nil
}

func (uc *HandleStripeWebhookUseCase) markFailed(ctx context.Context, p *domain.Payment) error {
	if p.Status == domain.StatusFailed {
		return nil // idempotent
	}
	p.Status = domain.StatusFailed
	p.UpdatedAt = time.Now().UTC()
	if err := uc.repo.Update(ctx, p); err != nil {
		return fmt.Errorf("aggiornamento pagamento fallito: %w", err)
	}
	_ = uc.publisher.Publish(ctx, events.SubjectPaymentFailed, events.PaymentFailed{
		TenantID:   p.TenantID,
		UserID:     p.UserID,
		PaymentID:  p.ID,
		Reason:     p.FailureReason,
		OccurredAt: p.UpdatedAt,
	})
	return nil
}
