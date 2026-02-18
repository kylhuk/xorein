package signaling

import (
	"testing"
)

func TestLifecycleTransitions(t *testing.T) {
	l := NewLifecycle(2)

	if err := l.CreateRoom(); err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if err := l.JoinRoom(); err != nil {
		t.Fatalf("unexpected join error: %v", err)
	}

	if err := l.LeaveRoom(); err != nil {
		t.Fatalf("unexpected leave error: %v", err)
	}

	if l.CurrentState() != StateLeft {
		t.Fatalf("expected state %q got %q", StateLeft, l.CurrentState())
	}
}

func TestRetryBudget(t *testing.T) {
	l := NewLifecycle(2)
	if _, err := l.Retry(); err == nil {
		t.Fatalf("expected retry error when not created")
	}

	if err := l.CreateRoom(); err != nil {
		t.Fatalf("create expected nil, got %v", err)
	}

	for i := 1; i <= 2; i++ {
		if _, err := l.Retry(); err != nil {
			t.Fatalf("retry %d failed: %v", i, err)
		}
	}

	if _, err := l.Retry(); err == nil {
		t.Fatalf("expected retry limit error after max attempts")
	}
}
