package daemon

import (
	"context"
	"fmt"
	"time"
)

// SupervisorRefusal signals the supervisor stopped retrying (bounded retries).
type SupervisorRefusal struct {
	Reason     string
	RetryAfter time.Duration
}

func (s *SupervisorRefusal) Error() string {
	return fmt.Sprintf("supervisor refusal: %s (retry-after=%s)", s.Reason, s.RetryAfter)
}

// SupervisorConfig controls retry limits and backoff.
type SupervisorConfig struct {
	MaxRetries int
	Backoff    time.Duration
}

// Supervisor restarts the daemon up to MaxRetries with exponential backoff.
type Supervisor struct {
	config SupervisorConfig
}

// NewSupervisor returns a supervisor with at least one retry configured.
func NewSupervisor(cfg SupervisorConfig) *Supervisor {
	if cfg.MaxRetries < 1 {
		cfg.MaxRetries = 3
	}

	if cfg.Backoff <= 0 {
		cfg.Backoff = 500 * time.Millisecond
	}

	return &Supervisor{config: cfg}
}

// Run attempts to start the daemon by executing startFn repeatedly.
func (s *Supervisor) Run(ctx context.Context, startFn func() error) error {
	backoff := s.config.Backoff
	attempt := 0

	for {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			backoff *= 2
		}

		err := startFn()
		if err == nil {
			return nil
		}

		attempt++
		if attempt >= s.config.MaxRetries {
			return &SupervisorRefusal{
				Reason:     "max retries exceeded",
				RetryAfter: backoff,
			}
		}
	}
}
