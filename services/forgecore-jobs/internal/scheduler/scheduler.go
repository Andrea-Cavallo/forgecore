package scheduler

import (
	"context"
	"log/slog"
	"time"
)

// Task represents a scheduled recurring task.
type Task struct {
	Name     string
	Interval time.Duration
	Run      func(ctx context.Context) error
}

// Scheduler runs tasks at fixed intervals.
type Scheduler struct {
	tasks []Task
}

func NewScheduler() *Scheduler {
	return &Scheduler{}
}

func (s *Scheduler) Register(t Task) {
	s.tasks = append(s.tasks, t)
}

func (s *Scheduler) Start(ctx context.Context) {
	for _, task := range s.tasks {
		t := task
		go s.run(ctx, t)
	}
}

func (s *Scheduler) run(ctx context.Context, t Task) {
	ticker := time.NewTicker(t.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("scheduler fermato", "task", t.Name)
			return
		case <-ticker.C:
			if err := t.Run(ctx); err != nil {
				slog.Error("task schedulato fallito", "task", t.Name, "errore", err)
			}
		}
	}
}
