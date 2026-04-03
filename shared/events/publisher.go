package events

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

// Publisher wraps a JetStream connection to publish typed events.
type Publisher struct{ js jetstream.JetStream }

// NewPublisher creates a Publisher backed by the given JetStream context.
func NewPublisher(js jetstream.JetStream) *Publisher { return &Publisher{js: js} }

// Publish marshals payload to JSON and publishes it to the given NATS subject.
func (p *Publisher) Publish(ctx context.Context, subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event %s: %w", subject, err)
	}
	if _, err = p.js.Publish(ctx, subject, data); err != nil {
		return fmt.Errorf("publish %s: %w", subject, err)
	}
	return nil
}
