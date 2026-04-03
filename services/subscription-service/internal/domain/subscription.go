package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive   = "active"
	StatusCanceled = "canceled"
	StatusPastDue  = "past_due"
	StatusTrialing = "trialing"

	IntervalMonthly = "monthly"
	IntervalYearly  = "yearly"
)

type Plan struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Name       string
	Amount     int64
	Currency   string
	Interval   string
	ProviderID string
	Active     bool
	CreatedAt  time.Time
}

type Subscription struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	UserID             uuid.UUID
	PlanID             uuid.UUID
	Status             string
	ProviderID         string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CanceledAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (s *Subscription) IsActive() bool {
	return s.Status == StatusActive || s.Status == StatusTrialing
}

func (s *Subscription) IsCancelable() bool {
	return s.Status == StatusActive || s.Status == StatusTrialing
}
