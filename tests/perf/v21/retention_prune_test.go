package v21

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v21/store"
	"github.com/aether/code_aether/pkg/v21/store/retention"
)

func TestRetentionPruneDeterministicBounds(t *testing.T) {
	policy := &retention.RetentionPolicy{WindowDays: 2, MaxEntries: 3}
	now := time.Date(2026, time.February, 18, 12, 0, 0, 0, time.UTC)
	events := []store.TimelineEvent{}
	addChannelEvents := func(channel string, count int) {
		for i := 0; i < count; i++ {
			events = append(events, store.TimelineEvent{
				Canonical: store.CanonicalEvent{
					MessageID: fmt.Sprintf("%s-%d", channel, i),
					EventID:   fmt.Sprintf("%s-%d", channel, i),
					ChannelID: channel,
					SenderID:  fmt.Sprintf("sender-%d", i),
				},
				Timestamp: now.AddDate(0, 0, -i),
				ChannelID: channel,
				Payload:   fmt.Sprintf("payload-%s", channel),
			})
		}
	}
	addChannelEvents("general", 6)
	addChannelEvents("alerts", 4)
	events = append(events, store.TimelineEvent{
		Canonical: store.CanonicalEvent{MessageID: "pinned-old", EventID: "pinned-old", ChannelID: "general", SenderID: "system"},
		Timestamp: now.AddDate(0, 0, -10),
		ChannelID: "general",
		Pinned:    true,
		Payload:   "pinned",
	})
	events = append(events, store.TimelineEvent{
		Canonical: store.CanonicalEvent{MessageID: "system-alert", EventID: "system-alert", ChannelID: "alerts", SenderID: "system"},
		Timestamp: now.AddDate(0, 0, -5),
		ChannelID: "alerts",
		System:    true,
		Payload:   "system",
	})
	windowStart := now.AddDate(0, 0, -policy.WindowDays)
	channelBuckets := map[string][]store.TimelineEvent{}
	for _, event := range events {
		channelBuckets[event.ChannelID] = append(channelBuckets[event.ChannelID], event)
	}
	expectedPruned := map[string]struct{}{}
	for channel, bucket := range channelBuckets {
		ordered := store.OrderedEvents(bucket)
		keptCount := 0
		for _, event := range ordered {
			if event.Pinned || event.System {
				continue
			}
			key := event.Canonical.Key()
			if !event.Timestamp.IsZero() && event.Timestamp.Before(windowStart) {
				expectedPruned[key] = struct{}{}
				continue
			}
			if policy.MaxEntries > 0 && keptCount >= policy.MaxEntries {
				expectedPruned[key] = struct{}{}
				continue
			}
			keptCount++
		}
		_ = channel // ensure channel variable used
	}
	result := policy.Prune(events, now)
	again := policy.Prune(events, now)
	if !reflect.DeepEqual(result, again) {
		t.Fatalf("retention result not deterministic: %+v vs %+v", result, again)
	}
	if len(result.Kept)+len(result.Pruned) != len(events) {
		t.Fatalf("expected kept+pruned to cover all events, got %d vs %d", len(result.Kept)+len(result.Pruned), len(events))
	}
	channelCounts := map[string]int{}
	for _, kept := range result.Kept {
		if kept.Pinned || kept.System {
			continue
		}
		if kept.Timestamp.Before(windowStart) {
			t.Fatalf("event %s should not survive when outside window", kept.Canonical.Key())
		}
		channelCounts[kept.ChannelID]++
		if channelCounts[kept.ChannelID] > policy.MaxEntries {
			t.Fatalf("channel %s exceeded max entries: %d", kept.ChannelID, channelCounts[kept.ChannelID])
		}
	}
	for _, pruned := range result.Pruned {
		if pruned.Pinned || pruned.System {
			t.Fatalf("pinned/system event was pruned: %s", pruned.Canonical.Key())
		}
	}
	actualPruned := map[string]struct{}{}
	for _, pruned := range result.Pruned {
		actualPruned[pruned.Canonical.Key()] = struct{}{}
	}
	if len(actualPruned) != len(expectedPruned) {
		t.Fatalf("expected %d pruned events, got %d", len(expectedPruned), len(actualPruned))
	}
	for key := range expectedPruned {
		if _, ok := actualPruned[key]; !ok {
			t.Fatalf("expected event %s to be pruned", key)
		}
	}
}
