package ingest

import (
	"reflect"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v21/store"
)

func canonical(id string) store.CanonicalEvent {
	return store.CanonicalEvent{
		MessageID: id,
		EventID:   id + "-v1",
		ChannelID: "chan",
		SenderID:  "sender",
	}
}

func TestApplyDeduplication(t *testing.T) {
	ing := NewIngestor(10)
	event := store.TimelineEvent{Canonical: canonical("msg-a"), Timestamp: time.Unix(1, 0), Payload: "one"}

	first := ing.Apply(event)
	if !first.Success || !first.Stored || first.Reason != "" {
		t.Fatalf("expected first insert to succeed, got %+v", first)
	}

	second := ing.Apply(event)
	if !second.Success || second.Stored {
		t.Fatalf("expected duplicate to succeed without storing, got %+v", second)
	}
}

func TestApplyFailureReasons(t *testing.T) {
	ing := NewIngestor(1)
	event := store.TimelineEvent{Canonical: canonical("msg-a"), Timestamp: time.Unix(1, 0), Payload: "one"}
	if !ing.Apply(event).Success {
		t.Fatalf("unexpected failure for first insert")
	}

	other := store.TimelineEvent{Canonical: canonical("msg-b"), Timestamp: time.Unix(2, 0), Payload: "two"}
	if res := ing.Apply(other); res.Reason != store.StoreQuotaExceeded {
		t.Fatalf("expected quota exceeded, got %v", res.Reason)
	}

	ing.Lock()
	if res := ing.Apply(event); res.Reason != store.StoreLocked {
		t.Fatalf("expected locked, got %v", res.Reason)
	}
	ing.Unlock()

	ing.MarkCorrupt(true)
	if res := ing.Apply(event); res.Reason != store.StoreCorrupt {
		t.Fatalf("expected corrupt, got %v", res.Reason)
	}
	ing.MarkCorrupt(false)

	ing.RequireMigration(true)
	if res := ing.Apply(event); res.Reason != store.StoreMigrationRequired {
		t.Fatalf("expected migration, got %v", res.Reason)
	}
}

func TestStoredEventsOrderingStable(t *testing.T) {
	ing := NewIngestor(10)
	now := time.Unix(1, 0)
	events := []store.TimelineEvent{
		{Canonical: canonical("first"), Timestamp: now, Payload: "first"},
		{Canonical: canonical("second"), Timestamp: now.Add(time.Second), Payload: "second"},
		{Canonical: canonical("third"), Timestamp: now.Add(2 * time.Second), Payload: "third"},
	}
	for _, event := range events {
		if res := ing.Apply(event); !res.Success || !res.Stored {
			t.Fatalf("unexpected insert result %v", res)
		}
	}
	// replay duplicates to ensure ordering does not change
	ing.Apply(events[1])
	ing.Apply(events[0])
	want := []string{"first", "second", "third"}
	if got := payloads(ing.StoredEvents()); !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected stored order: %v", got)
	}
	if got := payloads(ing.StoredEvents()); !reflect.DeepEqual(got, want) {
		t.Fatalf("expected repeated read to match, got %v", got)
	}
}

func payloads(events []store.TimelineEvent) []string {
	result := make([]string, len(events))
	for i, event := range events {
		result[i] = event.Payload
	}
	return result
}
