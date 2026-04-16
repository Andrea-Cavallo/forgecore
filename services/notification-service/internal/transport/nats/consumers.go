// Package nats subscribes to NATS JetStream subjects and dispatches to application use cases.
package nats

import (
	"context"
	"encoding/json"
	"log/slog"

	natsclient "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/Andrea-Cavallo/golang-modules/services/notification-service/internal/application"
	"github.com/Andrea-Cavallo/golang-modules/shared/events"
)

const (
	streamName    = "NOTIFICATIONS"
	consumerName  = "notification-service"
)

// Consumers holds the NATS JetStream consumers for the notification service.
type Consumers struct {
	js   jetstream.JetStream
	send *application.SendUseCase
}

// NewConsumers creates the consumers struct.
func NewConsumers(nc *natsclient.Conn, send *application.SendUseCase) (*Consumers, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}
	return &Consumers{js: js, send: send}, nil
}

// Start creates the durable consumer and starts message processing.
// It blocks until ctx is cancelled.
func (c *Consumers) Start(ctx context.Context) error {
	stream, err := c.ensureStream(ctx)
	if err != nil {
		return err
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       consumerName,
		FilterSubject: events.SubjectNotificationRequested,
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

func (c *Consumers) handleMessage(msg jetstream.Msg) {
	var ev events.NotificationRequested
	if err := json.Unmarshal(msg.Data(), &ev); err != nil {
		slog.Error("deserializzazione evento notifica fallita", "errore", err)
		_ = msg.Nak()
		return
	}

	ctx := context.Background()
	err := c.send.Execute(ctx, application.SendInput{
		TenantID:  ev.TenantID,
		UserID:    ev.UserID,
		Channel:   ev.Channel,
		Template:  ev.Template,
		Recipient: ev.Vars["recipient"],
		Vars:      ev.Vars,
	})
	if err != nil {
		slog.Error("invio notifica fallito", "errore", err, "tenant", ev.TenantID)
		_ = msg.Nak()
		return
	}
	_ = msg.Ack()
}

func (c *Consumers) ensureStream(ctx context.Context) (jetstream.Stream, error) {
	return c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{"notification.>"},
	})
}
