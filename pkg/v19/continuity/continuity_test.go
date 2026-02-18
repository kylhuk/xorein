package continuity

import "testing"

func TestTransitionSequence(t *testing.T) {
	next, reason := Transition(StateDraft, EventReadPersist)
	if next != StateRead || reason != "read-synced" {
		t.Fatalf("expected read state, got %v reason %v", next, reason)
	}

	next, reason = Transition(StateRead, EventCallHandoff)
	if next != StateCall || reason != "handoff-complete" {
		t.Fatalf("expected call state, got %v reason %v", next, reason)
	}

	next, reason = Transition(StateCall, EventWakeSignal)
	if next != StateWake || reason != "wake-ack" {
		t.Fatalf("expected wake state, got %v reason %v", next, reason)
	}
}
