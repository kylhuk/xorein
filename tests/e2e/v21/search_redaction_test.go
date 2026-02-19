package v21

import (
	"context"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v21/search"
	"github.com/aether/code_aether/pkg/v21/store/tombstone"
)

func TestRedactionKeepsSearchClean(t *testing.T) {
	idx := search.NewIndex()
	now := time.Date(2026, time.February, 18, 10, 0, 0, 0, time.UTC)
	idx.Add(search.Document{ID: "m1", Channel: "general", Sender: "alice", Timestamp: now, Body: "hello"})
	hooks := 0
	store := tombstone.NewStore(func(id string) {
		hooks++
		idx.Remove(id)
	}, tombstone.WithTimeSource(func() time.Time { return now }))

	if _, err := store.Apply(context.Background(), "m1", map[string]string{"channel": "general"}, "mod", "audit"); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if hooks != 1 {
		t.Fatalf("expected hook once, got %d", hooks)
	}
	results, err := idx.Search(context.Background(), search.QueryOptions{Limit: 5})
	if err != nil {
		t.Fatalf("search err: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected zero hits after redaction, got %d", len(results))
	}
}

func TestChannelAndTimeScoping(t *testing.T) {
	idx := search.NewIndex()
	base := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	idx.Add(search.Document{ID: "g1", Channel: "general", Sender: "alice", Timestamp: base, Body: "g"})
	idx.Add(search.Document{ID: "g2", Channel: "general", Sender: "bob", Timestamp: base.Add(2 * time.Hour), Body: "g2"})
	idx.Add(search.Document{ID: "alerts1", Channel: "alerts", Sender: "system", Timestamp: base.Add(3 * time.Hour), Body: "alert"})

	opts := search.QueryOptions{Channel: "general", Since: base.Add(time.Hour), Limit: 10}
	res, err := idx.Search(context.Background(), opts)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(res) != 1 || res[0].ID != "g2" {
		t.Fatalf("unexpected scoped results: %v", res)
	}

	coverage, ok := idx.Coverage("alerts")
	if !ok || !coverage.Until.Equal(base.Add(3*time.Hour)) {
		t.Fatalf("coverage mismatch: %+v", coverage)
	}
}
