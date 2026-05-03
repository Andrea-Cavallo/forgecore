package jobs

import (
	"context"
	"fmt"
	"log/slog"
)

// Handler processes a single job.
type Handler interface {
	Handle(ctx context.Context, payload []byte) error
}

// Registry maps job types to their handlers.
type Registry struct {
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

func (r *Registry) Register(jobType string, h Handler) {
	r.handlers[jobType] = h
}

func (r *Registry) Dispatch(ctx context.Context, jobType string, payload []byte) error {
	h, ok := r.handlers[jobType]
	if !ok {
		return fmt.Errorf("handler non trovato per job: %s", jobType)
	}
	if err := h.Handle(ctx, payload); err != nil {
		slog.Error("esecuzione job fallita", "tipo", jobType, "errore", err)
		return fmt.Errorf("job %s fallito: %w", jobType, err)
	}
	return nil
}

