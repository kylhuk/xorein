package search

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSearchFiltersAndCoverage(t *testing.T) {
	idx := NewIndex()
	base := time.Date(2026, time.February, 18, 12, 0, 0, 0, time.UTC)
	docs := []Document{
		{ID: "m1", Channel: "general", Sender: "alice", Timestamp: base, Body: "hello"},
		{ID: "m2", Channel: "general", Sender: "bob", Timestamp: base.Add(time.Hour), Body: "hi"},
		{ID: "m3", Channel: "alerts", Sender: "system", Timestamp: base.Add(2 * time.Hour), Body: "watch"},
	}
	for _, doc := range docs {
		idx.Add(doc)
	}

	results, err := idx.Search(context.Background(), QueryOptions{Channel: "general", Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Sender != "alice" || results[1].Sender != "bob" {
		t.Fatalf("unexpected ordering %v", results)
	}

	coverage, ok := idx.Coverage("general")
	if !ok || coverage.Status != "COVERAGE_FULL" {
		t.Fatalf("expected channel coverage, got %+v", coverage)
	}
	if coverage.Until != base.Add(time.Hour) {
		t.Fatalf("wrong coverage window %v", coverage)
	}

	filtered, err := idx.Search(context.Background(), QueryOptions{Sender: "system", Limit: 5})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(filtered) != 1 || filtered[0].Channel != "alerts" {
		t.Fatalf("expected alerts only, got %v", filtered)
	}
}

func TestSearchLimitGuard(t *testing.T) {
	idx := NewIndex()
	for i := 0; i < 6; i++ {
		idx.Add(Document{ID: fmt.Sprintf("limit-%d", i), Channel: "general", Sender: "sam", Timestamp: time.Now(), Body: "beat"})
	}
	_, err := idx.Search(context.Background(), QueryOptions{Limit: 3})
	if err == nil {
		t.Fatalf("expected timeout error when limit exceeded")
	}
	if err != ErrSearchQueryTimeout {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchTimeoutGuard(t *testing.T) {
	idx := NewIndex()
	idx.Add(Document{ID: "x", Channel: "general", Sender: "sam", Timestamp: time.Now(), Body: "ping"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := idx.Search(ctx, QueryOptions{Limit: 10})
	if err != ErrSearchQueryTimeout {
		t.Fatalf("expected timeout cancellation, got %v", err)
	}
}
