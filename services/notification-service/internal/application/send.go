package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/Andrea-Cavallo/golang-modules/services/notification-service/internal/domain"
)

type SendInput struct {
	TenantID  uuid.UUID
	UserID    uuid.UUID
	Channel   string
	Template  string
	Recipient string
	Vars      map[string]string
}

func (i SendInput) Validate() error {
	if i.Channel != domain.ChannelEmail && i.Channel != domain.ChannelSMS {
		return fmt.Errorf("canale non valido: %s", i.Channel)
	}
	if i.Recipient == "" {
		return fmt.Errorf("destinatario obbligatorio")
	}
	return nil
}

type SendUseCase struct {
	repo  domain.NotificationRepository
	email domain.EmailProvider
	sms   domain.SMSProvider
}

func NewSendUseCase(repo domain.NotificationRepository, email domain.EmailProvider, sms domain.SMSProvider) *SendUseCase {
	return &SendUseCase{repo: repo, email: email, sms: sms}
}

func (uc *SendUseCase) Execute(ctx context.Context, input SendInput) error {
	if err := input.Validate(); err != nil {
		return fmt.Errorf("input non valido: %w", err)
	}
	now := time.Now().UTC()
	n := &domain.Notification{
		ID:        uuid.New(),
		TenantID:  input.TenantID,
		UserID:    input.UserID,
		Channel:   input.Channel,
		Template:  input.Template,
		Recipient: input.Recipient,
		Vars:      input.Vars,
		Status:    domain.StatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}
	sendErr := uc.send(ctx, n)
	if sendErr != nil {
		n.Status = domain.StatusFailed
	} else {
		n.Status = domain.StatusSent
		n.SentAt = &now
	}
	n.Attempts++
	n.UpdatedAt = time.Now().UTC()
	if err := uc.repo.Create(ctx, n); err != nil {
		return fmt.Errorf("salvataggio notifica fallito: %w", err)
	}
	return sendErr
}

func (uc *SendUseCase) send(ctx context.Context, n *domain.Notification) error {
	switch n.Channel {
	case domain.ChannelEmail:
		return uc.email.Send(ctx, n.Recipient, n.Template, n.Vars)
	case domain.ChannelSMS:
		return uc.sms.Send(ctx, n.Recipient, n.Template, n.Vars)
	default:
		return domain.ErrInvalidChannel
	}
}
