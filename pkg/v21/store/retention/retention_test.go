package retention

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v21/store"
)

func timelineEvent(id string, daysAgo int, pinned bool) store.TimelineEvent {
	now := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	return store.TimelineEvent{
		Canonical: store.CanonicalEvent{MessageID: id, EventID: fmt.Sprintf("%s-v1", id), ChannelID: "chan", SenderID: "sender"},
		Timestamp: now.AddDate(0, 0, -daysAgo),
		Payload:   id,
		ChannelID: "chan",
		Pinned:    pinned,
	}
}

func channelEvent(id, channel string, daysAgo int, pinned bool) store.TimelineEvent {
	now := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	return store.TimelineEvent{
		Canonical: store.CanonicalEvent{MessageID: id, EventID: fmt.Sprintf("%s-v1", id), ChannelID: channel, SenderID: "sender"},
		Timestamp: now.AddDate(0, 0, -daysAgo),
		Payload:   id,
		ChannelID: channel,
		Pinned:    pinned,
	}
}

func TestPruneWindowAndLimit(t *testing.T) {
	rp := NewRetentionPolicy()
	rp.MaxEntries = 1
	now := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	events := []store.TimelineEvent{
		timelineEvent("old", 40, false),
		timelineEvent("recent", 5, false),
		timelineEvent("pinned", 60, true),
	}

	result := rp.Prune(events, now)

	if !containsPayload(result.Pruned, "old") {
		t.Fatalf("expected old event to be pruned, got kept=%v", result.Kept)
	}
	if !containsPayload(result.Kept, "recent") {
		t.Fatalf("expected recent event to be kept, got %v", result.Kept)
	}
	if !containsPayload(result.Kept, "pinned") {
		t.Fatalf("expected pinned event to survive, got %v", result.Kept)
	}
}

func TestPruneDeterministicAcrossChannelOrders(t *testing.T) {
	rp := NewRetentionPolicy()
	rp.MaxEntries = 1
	now := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)

	base := []store.TimelineEvent{
		channelEvent("alpha-old", "alpha", 40, false),
		channelEvent("alpha-new", "alpha", 2, false),
		channelEvent("beta-old", "beta", 35, false),
		channelEvent("beta-new", "beta", 1, false),
	}
	alt := []store.TimelineEvent{base[2], base[3], base[0], base[1]}

	first := rp.Prune(base, now)
	second := rp.Prune(alt, now)
	if !reflect.DeepEqual(payloads(first.Kept), payloads(second.Kept)) {
		t.Fatalf("kept slices diverged: %v vs %v", payloads(first.Kept), payloads(second.Kept))
	}
	if !reflect.DeepEqual(payloads(first.Pruned), payloads(second.Pruned)) {
		t.Fatalf("pruned slices diverged: %v vs %v", payloads(first.Pruned), payloads(second.Pruned))
	}
	wantKept := []string{"alpha-new", "beta-new"}
	if !reflect.DeepEqual(payloads(first.Kept), wantKept) {
		t.Fatalf("unexpected kept order: %v", payloads(first.Kept))
	}
	wantPruned := []string{"alpha-old", "beta-old"}
	if !reflect.DeepEqual(payloads(first.Pruned), wantPruned) {
		t.Fatalf("unexpected pruned order: %v", payloads(first.Pruned))
	}
}

func containsPayload(events []store.TimelineEvent, payload string) bool {
	for _, e := range events {
		if e.Payload == payload {
			return true
		}
	}
	return false
}

func payloads(events []store.TimelineEvent) []string {
	result := make([]string, len(events))
	for i, event := range events {
		result[i] = event.Payload
	}
	return result
}
