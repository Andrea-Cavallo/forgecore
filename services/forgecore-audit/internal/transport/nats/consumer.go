// Package nats subscribes to the wildcard audit.> subject and persists every event.
package nats

import (
	"context"
	"encoding/json"
	"log/slog"

	natsclient "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/Andrea-Cavallo/golang-modules/services/forgecore-audit/internal/application"
	"github.com/Andrea-Cavallo/golang-modules/shared/events"
)

const (
	auditStream   = "AUDIT"
	auditConsumer = "forgecore-audit"
)

// Consumer subscribes to all audit.> subjects and records them.
type Consumer struct {
	js     jetstream.JetStream
	record *application.RecordEventUseCase
}

// NewConsumer constructs the NATS consumer.
func NewConsumer(nc *natsclient.Conn, record *application.RecordEventUseCase) (*Consumer, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}
	return &Consumer{js: js, record: record}, nil
}

// Start creates the stream + durable consumer and blocks until ctx is cancelled.
func (c *Consumer) Start(ctx context.Context) error {
	stream, err := c.ensureStream(ctx)
	if err != nil {
		return err
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       auditConsumer,
		FilterSubject: events.SubjectAuditWildcard,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return err
	}

	cc, err := consumer.Consume(c.handleMessage)
	if err != nil {
		return err
	}
	defer cc.Stop()

	<-ctx.Done()
	return nil
}

func (c *Consumer) handleMessage(msg jetstream.Msg) {
	var ev events.AuditEvent
	if err := json.Unmarshal(msg.Data(), &ev); err != nil {
		slog.Error("deserializzazione evento audit fallita", "errore", err)
		_ = msg.Nak()
		return
	}

	ctx := context.Background()
	if err := c.record.Execute(ctx, application.RecordEventInput{
		TenantID:     ev.TenantID,
		ActorID:      ev.ActorID,
		ActorType:    ev.ActorType,
		Action:       ev.Action,
		ResourceID:   ev.ResourceID,
		ResourceType: ev.ResourceType,
		IPAddress:    ev.IPAddress,
		Metadata:     ev.Metadata,
		OccurredAt:   ev.OccurredAt,
	}); err != nil {
		slog.Error("registrazione evento audit fallita", "errore", err)
		_ = msg.Nak()
		return
	}
	_ = msg.Ack()
}

func (c *Consumer) ensureStream(ctx context.Context) (jetstream.Stream, error) {
	return c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     auditStream,
		Subjects: []string{"audit.>"},
	})
}
