package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Andrea-Cavallo/golang-modules/services/notification-service/internal/domain"
)

const (
	maxAttempts = 5
)

// retryDelays defines exponential back-off intervals per attempt index (0-based).
var retryDelays = []time.Duration{
	1 * time.Minute,
	5 * time.Minute,
	15 * time.Minute,
	1 * time.Hour,
	4 * time.Hour,
}

// RetryUseCase retries failed notifications up to maxAttempts times.
type RetryUseCase struct {
	repo domain.NotificationRepository
	send *SendUseCase
}

// NewRetryUseCase constructs the retry use case.
func NewRetryUseCase(repo domain.NotificationRepository, send *SendUseCase) *RetryUseCase {
	return &RetryUseCase{repo: repo, send: send}
}

// Execute fetches all failed notifications eligible for retry and re-sends them.
func (uc *RetryUseCase) Execute(ctx context.Context) error {
	pending, err := uc.repo.ListPendingRetries(ctx, maxAttempts)
	if err != nil {
		return fmt.Errorf("lettura notifiche in attesa di retry: %w", err)
	}
	for _, n := range pending {
		if err := uc.retryOne(ctx, n); err != nil {
			slog.Error("retry notifica fallito", "id", n.ID, "errore", err)
		}
	}
	return nil
}

func (uc *RetryUseCase) retryOne(ctx context.Context, n *domain.Notification) error {
	input := SendInput{
		TenantID:  n.TenantID,
		UserID:    n.UserID,
		Channel:   n.Channel,
		Template:  n.Template,
		Recipient: n.Recipient,
		Vars:      n.Vars,
	}
	sendErr := uc.send.Execute(ctx, input)
	if sendErr != nil {
		n.Status = domain.StatusFailed
	} else {
		now := time.Now().UTC()
		n.Status = domain.StatusSent
		n.SentAt = &now
	}
	n.Attempts++
	n.UpdatedAt = time.Now().UTC()
	if err := uc.repo.Update(ctx, n); err != nil {
		return fmt.Errorf("aggiornamento retry notifica: %w", err)
	}
	return sendErr
}

// NextRetryDelay returns the delay before the next retry attempt (0-based index).
func NextRetryDelay(attempt int) time.Duration {
	if attempt < len(retryDelays) {
		return retryDelays[attempt]
	}
	return retryDelays[len(retryDelays)-1]
}
