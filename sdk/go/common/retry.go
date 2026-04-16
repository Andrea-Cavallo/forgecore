// Package common fornisce utility condivise tra tutti i client SDK.
package common

import (
	"context"
	"fmt"
	"time"
)

const (
	maxRetries    = 3
	retryInterval = 200 * time.Millisecond
)

// RetryFn esegue fn con retry esponenziale (max 3 tentativi).
func RetryFn(ctx context.Context, fn func() error) error {
	var lastErr error
	for attempt := range maxRetries {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("contesto annullato prima del tentativo %d: %w", attempt+1, err)
		}
		lastErr = fn()
		if lastErr == nil {
			return nil
		}
		if attempt < maxRetries-1 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("contesto annullato: %w", ctx.Err())
			case <-time.After(retryInterval * time.Duration(1<<attempt)):
			}
		}
	}
	return fmt.Errorf("tutti i %d tentativi falliti: %w", maxRetries, lastErr)
}
