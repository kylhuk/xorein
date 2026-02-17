package notification

import (
	"testing"
	"time"
)

func TestPipelineSuppressionAndAction(t *testing.T) {
	pipe := NewPipeline(30 * time.Second)
	now := time.Date(2025, 1, 1, 5, 0, 0, 0, time.UTC)
	notif := Notification{ServerID: "srv", ChannelID: "chan", MessageID: "msg", Trigger: TriggerDesktop}
	action, status := pipe.Fire(notif, now)
	if status != "notification.desktop.ready" {
		t.Fatalf("expected ready status, got %s", status)
	}
	if action.Label != "server:srv" {
		t.Fatalf("unexpected action label %s", action.Label)
	}
	second, suppressed := pipe.Fire(notif, now.Add(10*time.Second))
	if suppressed != "notification.desktop.suppressed" {
		t.Fatalf("expected suppression status")
	}
	if second.Trigger != TriggerDesktop {
		t.Fatalf("unexpected trigger in suppressed action")
	}
}

func TestParseActionTarget(t *testing.T) {
	target := ParseActionTarget("srv:chan:msg")
	if target.ServerID != "srv" || target.ChannelID != "chan" || target.MessageID != "msg" {
		t.Fatalf("unexpected action target %v", target)
	}
	target = ParseActionTarget("srv")
	if target.ChannelID != "" {
		t.Fatalf("expected empty channel id")
	}
}

func TestPipelineDedupeWindow(t *testing.T) {
	window := DedupeWindow{LastFired: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Interval: 15 * time.Second}
	if !window.ShouldSuppress(window.LastFired.Add(5 * time.Second)) {
		t.Fatalf("expected suppression within interval")
	}
	if window.ShouldSuppress(window.LastFired.Add(20 * time.Second)) {
		t.Fatalf("expected no suppression after interval")
	}
}

func TestNotificationPayloadTargetAndTrigger(t *testing.T) {
	if got := BuildPayloadTarget("srv", "chan", "msg", KindDM); got != "srv:chan:msg" {
		t.Fatalf("unexpected payload target %s", got)
	}
	if trigger := DetermineTrigger(KindCallInvite); trigger != TriggerPush {
		t.Fatalf("expected push trigger for call invite, got %s", trigger)
	}
	if trigger := DetermineTrigger(KindMention); trigger != TriggerDesktop {
		t.Fatalf("expected desktop trigger for mention, got %s", trigger)
	}
}
