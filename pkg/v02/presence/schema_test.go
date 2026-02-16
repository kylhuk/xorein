package presence

import (
	"testing"
	"time"
)

func TestPresenceTransitions(t *testing.T) {
	next, reason := NextState(StateOnline, EventIdleTimeout)
	if next != StateIdle || reason != ReasonTransitionApplied {
		t.Fatalf("idle transition mismatch: state=%q reason=%q", next, reason)
	}
	next, reason = NextState(StateDND, EventIdleTimeout)
	if next != StateDND || reason != ReasonIgnoredByPrecedence {
		t.Fatalf("dnd idle mismatch: state=%q reason=%q", next, reason)
	}
}

func TestValidateCustomStatus(t *testing.T) {
	status, err := ValidateCustomStatus("  hello  ", 16)
	if err != nil {
		t.Fatalf("validate status: %v", err)
	}
	if status != "hello" {
		t.Fatalf("status=%q want hello", status)
	}
	if _, err := ValidateCustomStatus("this is too long", 4); err == nil {
		t.Fatal("expected length error")
	}
}

func TestIsStale(t *testing.T) {
	now := time.Unix(100, 0)
	if !IsStale(now.Add(-11*time.Minute), now, 10*time.Minute) {
		t.Fatal("expected stale")
	}
	if IsStale(now.Add(-9*time.Minute), now, 10*time.Minute) {
		t.Fatal("expected fresh")
	}
}
