package screenshare

import "testing"

func TestTransition(t *testing.T) {
	state := Transition(StateIdle, EventStart)
	if state != StateConnecting {
		t.Fatalf("expected connecting, got %s", state)
	}
	state = Transition(state, EventConnected)
	if state != StateActive {
		t.Fatalf("expected active, got %s", state)
	}
	state = Transition(state, EventQualityDrop)
	if state != StateFallback {
		t.Fatalf("expected fallback, got %s", state)
	}
	state = Transition(state, EventRecover)
	if state != StateActive {
		t.Fatalf("expected active after recover, got %s", state)
	}
	state = Transition(state, EventQualityDrop)
	state = Transition(state, EventQualityDrop)
	if state != StateError {
		t.Fatalf("expected error after double drop, got %s", state)
	}
}

func TestDetermineAdaptation(t *testing.T) {
	low := DetermineAdaptation(1000)
	if low.Layer != 0 {
		t.Fatalf("expected layer0, got %d", low.Layer)
	}
	critical := DetermineAdaptation(500)
	if critical.Layer != 0 || critical.BitrateKbps != 600 {
		t.Fatalf("expected critical layer 0, got %v", critical)
	}
}

func TestLabel(t *testing.T) {
	desc := Summarize(StateActive, 3200)
	if desc.Decision.Layer != 2 {
		t.Fatalf("unexpected decision %v", desc.Decision)
	}
	label := Label(desc.State, desc.Decision)
	if label == "" {
		t.Fatal("label should not be empty")
	}
}
