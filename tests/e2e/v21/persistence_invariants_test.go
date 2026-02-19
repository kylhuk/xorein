package v21

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v21/store"
	"github.com/aether/code_aether/pkg/v21/store/hydrate"
	"github.com/aether/code_aether/pkg/v21/store/ingest"
)

func TestRestartSnapshotRestorePreservesOrdering(t *testing.T) {
	base := time.Date(2026, time.February, 18, 8, 0, 0, 0, time.UTC)
	set := []store.TimelineEvent{
		newEvent("m-alert", "alerts", "system", base.Add(-2*time.Hour)),
		newEvent("m-dup", "general", "alice", base),
		newEvent("m-bob", "general", "bob", base.Add(30*time.Minute)),
		newEvent("m-dup", "general", "alice", base.Add(1*time.Hour)),
		newEvent("m-norm", "general", "carol", base.Add(2*time.Hour)),
	}
	order := []int{0, 2, 1, 3, 4}
	ingestor := ingest.NewIngestor(0)
	for _, idx := range order {
		if res := ingestor.Apply(set[idx]); !res.Success {
			t.Fatalf("apply %d failed: %v", idx, res)
		}
	}
	snapshot := append([]store.TimelineEvent(nil), ingestor.StoredEvents()...)
	if len(snapshot) != 4 {
		t.Fatalf("expected 4 unique events after dedupe, got %d", len(snapshot))
	}
	expected := store.OrderedEvents(snapshot)
	h1 := &hydrate.Hydrator{}
	h1.SetEvents(snapshot)
	page := h1.Paginate("", len(snapshot))
	if !reflect.DeepEqual(page.Events, expected) {
		t.Fatalf("snapshot ordering mismatch: got %v want %v", page.Events, expected)
	}
	h2 := &hydrate.Hydrator{}
	h2.SetEvents(snapshot)
	h2.SetEvents(snapshot)
	page2 := h2.Paginate("", len(snapshot))
	if !reflect.DeepEqual(page2.Events, expected) {
		t.Fatalf("restore ordering mismatch: got %v want %v", page2.Events, expected)
	}
}

func TestIngestorMigrationRefusalDeterministic(t *testing.T) {
	ingestor := ingest.NewIngestor(5)
	ingestor.RequireMigration(true)
	event := newEvent("m-upgrade", "general", "dave", time.Date(2026, time.February, 18, 9, 0, 0, 0, time.UTC))
	for i := 0; i < 2; i++ {
		res := ingestor.Apply(event)
		if res.Reason != store.StoreMigrationRequired {
			t.Fatalf("iteration %d: expected migration failure reason, got %v", i, res.Reason)
		}
		if res.Stored {
			t.Fatalf("iteration %d: migrated store should not record events", i)
		}
	}
	ingestor.RequireMigration(false)
	first := ingestor.Apply(event)
	if !first.Stored || !first.Success {
		t.Fatalf("expected success after migration, got %v", first)
	}
	second := ingestor.Apply(event)
	if second.Stored {
		t.Fatalf("duplicate after migration should not store again")
	}
}

func TestIngestorCorruptionDetectionAndRepair(t *testing.T) {
	ingestor := ingest.NewIngestor(5)
	event := newEvent("m-fix", "alerts", "system", time.Date(2026, time.February, 18, 10, 0, 0, 0, time.UTC))
	ingestor.MarkCorrupt(true)
	if res := ingestor.Apply(event); res.Reason != store.StoreCorrupt || res.Success {
		t.Fatalf("corrupt state should refuse writes, got %v", res)
	}
	if len(ingestor.StoredEvents()) != 0 {
		t.Fatalf("corrupt store should not accumulate events")
	}
	ingestor.MarkCorrupt(false)
	res := ingestor.Apply(event)
	if !res.Stored || !res.Success {
		t.Fatalf("expected repair path to accept writes, got %v", res)
	}
	res = ingestor.Apply(event)
	if res.Stored {
		t.Fatalf("replay after repair should dedupe")
	}
}

func newEvent(id, channel, sender string, ts time.Time) store.TimelineEvent {
	return store.TimelineEvent{
		Canonical: store.CanonicalEvent{
			MessageID: id,
			EventID:   id,
			ChannelID: channel,
			SenderID:  sender,
		},
		Timestamp: ts,
		ChannelID: channel,
		Payload:   fmt.Sprintf("payload-%s", id),
	}
}
