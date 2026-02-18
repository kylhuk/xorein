package ui

import "testing"

func TestCallStateLabels(t *testing.T) {
	joined := CallState{Joined: true, Muted: true, Deafened: false, SelectedDevice: "headset"}
	if joined.ActionLabel() != "Leave" {
		t.Fatalf("expected leave label, got %q", joined.ActionLabel())
	}
	if joined.MuteLabel() != "Unmute" {
		t.Fatalf("expected unmute label")
	}
	if joined.DeafLabel() != "Deafen" {
		t.Fatalf("expected deafen label")
	}
}

func TestQualityBadgeNoLimbo(t *testing.T) {
	if QualityBadge(90) != "HD" {
		t.Fatalf("expected HD badge for 90")
	}
	if QualityBadge(65) != "SD" {
		t.Fatalf("expected SD badge for 65")
	}
	if QualityBadge(40) != "Degraded" {
		t.Fatalf("expected Degraded badge")
	}

	if msg := NoLimboMessage(70, false); msg != "Audio stable" {
		t.Fatalf("unexpected message: %s", msg)
	}
	if msg := NoLimboMessage(45, false); msg == "Audio stable" {
		t.Fatalf("expected degraded hint for 45")
	}
	if msg := NoLimboMessage(30, true); msg != "Reconnecting with recovery-first flow..." {
		t.Fatalf("expected recovery message")
	}
}
