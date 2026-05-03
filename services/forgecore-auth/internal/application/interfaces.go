package application

import "context"

// EventPublisher abstracts event publishing for testability.
type EventPublisher interface {
	Publish(ctx context.Context, subject string, payload any) error
}

// noopPublisher is a null implementation for tests or disabled environments.
type noopPublisher struct{}

func (noopPublisher) Publish(_ context.Context, _ string, _ any) error { return nil }
