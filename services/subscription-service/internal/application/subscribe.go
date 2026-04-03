package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/golang-modules/services/subscription-service/internal/domain"
)

type SubscribeInput struct {
	TenantID   uuid.UUID
	UserID     uuid.UUID
	PlanID     uuid.UUID
	CustomerID string
}

func (i SubscribeInput) Validate() error {
	if i.CustomerID == "" {
		return fmt.Errorf("customer ID obbligatorio")
	}
	return nil
}

type SubscribeOutput struct {
	Subscription *domain.Subscription
}

type SubscribeUseCase struct {
	subs    domain.SubscriptionRepository
	plans   domain.PlanRepository
	billing domain.BillingProvider
}

func NewSubscribeUseCase(subs domain.SubscriptionRepository, plans domain.PlanRepository, billing domain.BillingProvider) *SubscribeUseCase {
	return &SubscribeUseCase{subs: subs, plans: plans, billing: billing}
}

func (uc *SubscribeUseCase) Execute(ctx context.Context, input SubscribeInput) (*SubscribeOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("input non valido: %w", err)
	}
	plan, err := uc.plans.GetByID(ctx, input.PlanID, input.TenantID)
	if err != nil {
		return nil, fmt.Errorf("piano non trovato: %w", err)
	}
	existing, _ := uc.subs.GetActiveByUser(ctx, input.UserID, input.TenantID)
	if existing != nil {
		return nil, domain.ErrAlreadySubscribed
	}
	providerID, err := uc.billing.CreateSubscription(ctx, input.CustomerID, plan.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("creazione abbonamento provider fallita: %w", err)
	}
	now := time.Now().UTC()
	sub := &domain.Subscription{
		ID:                 uuid.New(),
		TenantID:           input.TenantID,
		UserID:             input.UserID,
		PlanID:             input.PlanID,
		Status:             domain.StatusActive,
		ProviderID:         providerID,
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := uc.subs.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("salvataggio abbonamento fallito: %w", err)
	}
	return &SubscribeOutput{Subscription: sub}, nil
}
