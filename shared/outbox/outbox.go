package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending   = "pending"
	StatusPublished = "published"
	StatusFailed    = "failed"
)

type Message struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	EventName     string
	EventVersion  int
	Subject       string
	Payload       json.RawMessage
	CorrelationID string
	Status        string
	Attempts      int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Repository interface {
	Save(ctx context.Context, msg Message) error
	FetchPending(ctx context.Context, limit int) ([]Message, error)
	MarkPublished(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, reason string) error
}

type Publisher interface {
	Publish(ctx context.Context, subject string, payload []byte) error
}

type Dispatcher struct {
	repo      Repository
	publisher Publisher
}

func NewDispatcher(repo Repository, publisher Publisher) *Dispatcher {
	return &Dispatcher{repo: repo, publisher: publisher}
}

func NewMessage(tenantID uuid.UUID, subject string, eventName string, version int, payload any, correlationID string) (Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Message{}, fmt.Errorf("serializzazione payload outbox: %w", err)
	}
	now := time.Now().UTC()
	return Message{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EventName:     eventName,
		EventVersion:  version,
		Subject:       subject,
		Payload:       data,
		CorrelationID: correlationID,
		Status:        StatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (d *Dispatcher) DispatchPending(ctx context.Context, limit int) error {
	messages, err := d.repo.FetchPending(ctx, limit)
	if err != nil {
		return fmt.Errorf("lettura outbox: %w", err)
	}
	for _, msg := range messages {
		if err := d.dispatch(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

func (d *Dispatcher) dispatch(ctx context.Context, msg Message) error {
	if err := d.publisher.Publish(ctx, msg.Subject, msg.Payload); err != nil {
		if markErr := d.repo.MarkFailed(ctx, msg.ID, err.Error()); markErr != nil {
			return fmt.Errorf("pubblicazione outbox fallita e mark failed fallito: %w", markErr)
		}
		return fmt.Errorf("pubblicazione outbox fallita: %w", err)
	}
	if err := d.repo.MarkPublished(ctx, msg.ID); err != nil {
		return fmt.Errorf("mark outbox published fallito: %w", err)
	}
	return nil
}
