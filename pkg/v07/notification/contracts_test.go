package notification

import (
	"testing"
	"time"
)

func TestDedupeWindowSuppression(t *testing.T) {
	window := DedupeWindow{LastFired: time.Now().Add(-time.Second), Interval: 5 * time.Second}
	if !window.ShouldSuppress(time.Now()) {
		t.Fatalf("expected suppression true when within interval")
	}
}

func TestResolveActionFallback(t *testing.T) {
	if ResolveAction(TriggerDesktop, "custom") != "custom" {
		t.Fatalf("expected fallback to win")
	}
	if ResolveAction(TriggerPush, "") != "push:show" {
		t.Fatalf("expected push default action")
	}
}

func TestSuppressionReason(t *testing.T) {
	reason := SuppressionReason(true, TriggerDesktop)
	if reason != string(TriggerDesktop)+".suppressed" {
		t.Fatalf("unexpected reason %s", reason)
	}
}
