package tombstone

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v21/search"
)

func TestTombstoneRemovesContentAndCallsHook(t *testing.T) {
	idx := search.NewIndex()
	now := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	idx.Add(search.Document{ID: "msg1", Channel: "general", Sender: "alice", Timestamp: now, Body: "hi"})
	calls := 0
	store := NewStore(func(id string) {
		calls++
		idx.Remove(id)
	}, WithTimeSource(func() time.Time { return now }))

	entry, err := store.Apply(context.Background(), "msg1", map[string]string{"channel": "general"}, "mod", "audit")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if entry.ContentExists {
		t.Fatalf("expected content removed")
	}
	if calls != 1 {
		t.Fatalf("expected hook called once, got %d", calls)
	}
	res, err := idx.Search(context.Background(), search.QueryOptions{Limit: 10})
	if err != nil {
		t.Fatalf("search err: %v", err)
	}
	if len(res) != 0 {
		t.Fatalf("expected redaction to remove search hits, got %d", len(res))
	}
}

func TestTombstoneIdempotent(t *testing.T) {
	now := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	store := NewStore(nil, WithTimeSource(func() time.Time { return now }))
	if _, err := store.Apply(context.Background(), "msg2", map[string]string{"channel": "general"}, "mod", "audit"); err != nil {
		t.Fatalf("apply1: %v", err)
	}
	if _, err := store.Apply(context.Background(), "msg2", map[string]string{"channel": "general"}, "mod", "audit"); err != nil {
		t.Fatalf("apply2: %v", err)
	}
	state := store.Snapshot()
	if len(state) != 1 {
		t.Fatalf("expected single state, got %d", len(state))
	}
	if state[0].TombstonedAt != now {
		t.Fatalf("expected stable tombstone time, got %v", state[0].TombstonedAt)
	}
}

func TestRestartInvariant(t *testing.T) {
	now := time.Date(2026, time.February, 18, 1, 0, 0, 0, time.UTC)
	store := NewStore(nil, WithTimeSource(func() time.Time { return now }))
	if _, err := store.Apply(context.Background(), "msg3", map[string]string{"channel": "general"}, "mod", "audit"); err != nil {
		t.Fatalf("apply: %v", err)
	}
	states := store.Snapshot()
	clone := NewStore(nil)
	clone.Restore(states)
	if !reflect.DeepEqual(states, clone.Snapshot()) {
		t.Fatalf("restart invariant violated")
	}
}

func TestMigrationAndCorruptionInvariants(t *testing.T) {
	store := NewStore(nil)
	bad := []EntryState{{ID: "ok"}, {ID: ""}}
	if err := store.Validate(bad); !errors.Is(err, ErrCorruptedEntry) {
		t.Fatalf("expected corruption error, got %v", err)
	}
	repaired := store.Repair(bad)
	if len(repaired) != 1 || repaired[0].ID != "ok" {
		t.Fatalf("repair failed, got %v", repaired)
	}
	clone := NewStore(nil)
	clone.Restore(repaired)
	if len(clone.Snapshot()) != 1 {
		t.Fatalf("restore failed after repair")
	}
}

func TestPruneConsistencyInvariant(t *testing.T) {
	now := time.Date(2026, time.February, 18, 0, 30, 0, 0, time.UTC)
	store := NewStore(nil, WithTimeSource(func() time.Time { return now }))
	if _, err := store.Apply(context.Background(), "old", map[string]string{"channel": "general"}, "prune", "audit"); err != nil {
		t.Fatalf("apply: %v", err)
	}
	store.PruneBefore(now.Add(time.Minute))
	if len(store.Snapshot()) != 0 {
		t.Fatalf("prune should remove tombstoned entry")
	}
}
