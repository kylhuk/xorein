package daemon

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSupervisorRetries(t *testing.T) {
	s := NewSupervisor(SupervisorConfig{MaxRetries: 3, Backoff: 1 * time.Millisecond})
	attempts := 0

	err := s.Run(context.Background(), func() error {
		attempts++
		if attempts < 2 {
			return errors.New("boom")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestSupervisorRefusal(t *testing.T) {
	s := NewSupervisor(SupervisorConfig{MaxRetries: 1, Backoff: 1 * time.Millisecond})

	err := s.Run(context.Background(), func() error {
		return errors.New("boom")
	})

	refusal, ok := err.(*SupervisorRefusal)
	if !ok {
		t.Fatalf("expected SupervisorRefusal, got %T", err)
	}

	if refusal.Reason != "max retries exceeded" {
		t.Fatalf("reason = %s", refusal.Reason)
	}
}
