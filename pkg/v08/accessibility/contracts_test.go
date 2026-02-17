package accessibility

import (
	"testing"
	"time"
)

func TestValidLabel(t *testing.T) {
	if !ValidLabel("button") {
		t.Fatalf("expected valid label")
	}
	if ValidLabel("unknown") {
		t.Fatalf("unexpectedly valid label")
	}
}

func TestAnnouncerThrottle(t *testing.T) {
	announcer := NewAnnouncer(1 * time.Second)
	now := time.Now()
	if !announcer.ShouldAnnounce("alert", now) {
		t.Fatalf("first announcement should pass")
	}
	if announcer.ShouldAnnounce("alert", now.Add(500*time.Millisecond)) {
		t.Fatalf("announcement should be throttled")
	}
	if !announcer.ShouldAnnounce("alert", now.Add(2*time.Second)) {
		t.Fatalf("announcement should succeed after limit")
	}
}

func TestHighContrastToken(t *testing.T) {
	if got := HighContrastToken(true, "#abc"); got != "#000000" {
		t.Fatalf("expected override, got %s", got)
	}
	if got := HighContrastToken(false, "#abc"); got != "#abc" {
		t.Fatalf("expected original, got %s", got)
	}
}

func TestFocusGraph(t *testing.T) {
	graph := FocusGraph{}
	graph.AddEdge("start", "next")
	graph.AddEdge("start", "alt")
	neigh := graph.Neighbors("start")
	if len(neigh) != 2 {
		t.Fatalf("expected two neighbors, got %d", len(neigh))
	}
}
