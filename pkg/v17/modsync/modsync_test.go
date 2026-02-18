package modsync_test

import (
	"testing"

	"github.com/aether/code_aether/pkg/v17/moderation"
	"github.com/aether/code_aether/pkg/v17/modsync"
)

func TestOrdererMerge(t *testing.T) {
	orderer := modsync.NewOrderer()
	base := []moderation.SignedEvent{
		{ID: "evt-a", Room: "room:1", Actor: "mod@relay", Target: "user:one", Type: moderation.EventTimeout, Timestamp: 10, Signature: "sig:mod@relay"},
		{ID: "evt-b", Room: "room:1", Actor: "mod@relay", Target: "user:two", Type: moderation.EventBan, Timestamp: 20, Signature: "sig:mod@relay"},
	}
	extra := []moderation.SignedEvent{
		{ID: "evt-c", Room: "room:1", Actor: "mod@relay", Target: "user:three", Type: moderation.EventLockdown, Timestamp: 15, Signature: "sig:mod@relay"},
		base[0],
	}
	merged := orderer.Merge(base, extra)
	if len(merged) != 3 {
		t.Fatalf("expected 3 unique events, got %d", len(merged))
	}
	if merged[0].ID != "evt-a" || merged[1].ID != "evt-c" || merged[2].ID != "evt-b" {
		t.Fatalf("unexpected order %v", []string{merged[0].ID, merged[1].ID, merged[2].ID})
	}
}

func TestOrdererKnown(t *testing.T) {
	orderer := modsync.NewOrderer()
	event := moderation.SignedEvent{ID: "evt-d", Room: "room:alpha", Actor: "mod@relay", Target: "user:x", Type: moderation.EventSlowMode, Timestamp: 30, Signature: "sig:mod@relay"}
	orderer.Merge([]moderation.SignedEvent{event})
	if !orderer.Known(event) {
		t.Fatalf("event should be known after merge")
	}
}
