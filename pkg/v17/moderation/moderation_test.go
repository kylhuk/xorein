package moderation_test

import (
	"testing"

	"github.com/aether/code_aether/pkg/v17/moderation"
)

func TestEngineAcceptsValidEvent(t *testing.T) {
	engine := moderation.NewEngine()
	event := moderation.SignedEvent{
		ID:        "evt-1",
		Room:      "room:alpha",
		Actor:     "mod@relay",
		Target:    "user:alice",
		Type:      moderation.EventBan,
		Timestamp: 100,
		Signature: "sig:mod@relay",
	}
	result := engine.Apply(event)
	if !result.Accepted {
		t.Fatalf("expected accept, got %v", result)
	}
	if reason := result.Reason; reason != moderation.ReasonAccepted {
		t.Fatalf("unexpected reason %s", reason)
	}
	if got := engine.SeenCount(); got != 1 {
		t.Fatalf("expected seen count 1, got %d", got)
	}
}

func TestEngineRejectsInvalidSignature(t *testing.T) {
	engine := moderation.NewEngine()
	event := moderation.SignedEvent{ID: "evt-2", Room: "room:alpha", Actor: "mod@relay", Target: "user:bob", Type: moderation.EventBan, Timestamp: 101, Signature: "bad"}
	result := engine.Apply(event)
	if result.Reason != moderation.ReasonInvalidSignature {
		t.Fatalf("expected invalid-signature reason, got %s", result.Reason)
	}
	if count := engine.SeenCount(); count != 0 {
		t.Fatalf("invalid signature should not mark seen events, got %d", count)
	}
}

func TestEngineRejectsDuplicatesAndStale(t *testing.T) {
	engine := moderation.NewEngine()
	first := moderation.SignedEvent{ID: "evt-3", Room: "room:beta", Actor: "mod@relay", Target: "user:carol", Type: moderation.EventRedaction, Timestamp: 200, Signature: "sig:mod@relay"}
	if res := engine.Apply(first); !res.Accepted {
		t.Fatalf("unexpected reject %v", res)
	}
	duplicate := first
	second := moderation.SignedEvent{ID: "evt-4", Room: "room:beta", Actor: "mod@relay", Target: "user:dan", Type: moderation.EventTimeout, Timestamp: 200, Signature: "sig:mod@relay"}
	if res := engine.Apply(duplicate); res.Reason != moderation.ReasonDuplicateEvent {
		t.Fatalf("expected duplicate, got %v", res)
	}
	if res := engine.Apply(second); res.Reason != moderation.ReasonStaleEvent {
		t.Fatalf("expected stale, got %v", res)
	}
}

func TestSequenceTraceDeterministic(t *testing.T) {
	engine := moderation.NewEngine()
	events := []moderation.SignedEvent{
		{ID: "evt-5", Room: "room:gamma", Actor: "mod@relay", Target: "user:erin", Type: moderation.EventSlowMode, Timestamp: 10, Signature: "sig:mod@relay"},
		{ID: "evt-6", Room: "room:gamma", Actor: "mod@relay", Target: "user:frank", Type: moderation.EventLockdown, Timestamp: 20, Signature: "sig:mod@relay"},
	}
	for _, evt := range events {
		if res := engine.Apply(evt); !res.Accepted {
			t.Fatalf("unexpected reject %v", res)
		}
	}
	trace := engine.SequenceTrace()
	if len(trace) != 2 {
		t.Fatalf("expected trace len 2, got %d", len(trace))
	}
	if trace[0] == trace[1] {
		t.Fatalf("trace entries should be distinct, got %v", trace)
	}
}
