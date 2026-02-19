package daemon

import (
	"testing"
)

func TestLockManagerAcquireRelease(t *testing.T) {
	mgr := NewLockManager("/tmp/xorein-lock")

	lock, err := mgr.Acquire("prime", StateStarting)
	if err != nil {
		t.Fatalf("unexpected acquire error: %v", err)
	}

	if got := mgr.State(); got != StateStarting {
		t.Fatalf("state = %s, want %s", got, StateStarting)
	}

	if err := lock.Release(StateRunning); err != nil {
		t.Fatalf("release failed: %v", err)
	}

	if got := mgr.State(); got != StateRunning {
		t.Fatalf("state after release = %s, want %s", got, StateRunning)
	}
}

func TestLockManagerBusyRefusal(t *testing.T) {
	mgr := NewLockManager("/tmp/xorein-lock2")

	lock, err := mgr.Acquire("primary", StateStarting)
	if err != nil {
		t.Fatalf("unexpected inaugural acquire: %v", err)
	}

	if _, err := mgr.Acquire("secondary", StateRunning); err == nil {
		t.Fatalf("expected busy refusal")
	} else if refusal, ok := err.(*RefusalError); !ok {
		t.Fatalf("expected RefusalError, got %T", err)
	} else if refusal.Reason != LockUnavailable {
		t.Fatalf("reason = %s, want %s", refusal.Reason, LockUnavailable)
	}

	if err := lock.Release(StateRunning); err != nil {
		t.Fatalf("release: %v", err)
	}
}

func TestLockManagerInvalidTransition(t *testing.T) {
	mgr := NewLockManager("/tmp/xorein-lock3")

	if _, err := mgr.Acquire("prime", StateRunning); err == nil {
		t.Fatalf("expected transition refusal")
	} else if refusal, ok := err.(*RefusalError); !ok {
		t.Fatalf("expected RefusalError, got %T", err)
	} else if refusal.Reason != LockInvalidTransition {
		t.Fatalf("reason = %s, want %s", refusal.Reason, LockInvalidTransition)
	}

	lock, err := mgr.Acquire("prime", StateStarting)
	if err != nil {
		t.Fatalf("fallback acquire: %v", err)
	}

	if err := lock.Release(StateStopping); err == nil {
		t.Fatalf("expected invalid release transition")
	}

	if err := lock.Release(StateRunning); err != nil {
		t.Fatalf("release after fix: %v", err)
	}
}
