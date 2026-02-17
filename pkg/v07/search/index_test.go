package search

import (
	"testing"
	"time"
)

func TestIndexRebuildDeterministic(t *testing.T) {
	idx := NewIndex()
	docs := []Document{
		{ID: "d1", Scope: "scope", CreatedAt: time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)},
		{ID: "d2", Scope: "scope", CreatedAt: time.Date(2025, 1, 1, 9, 0, 0, 0, time.UTC)},
	}
	for _, doc := range docs {
		if err := idx.Insert(doc); err != nil {
			t.Fatalf("insert failed: %v", err)
		}
	}
	first := idx.Rebuild()
	second := idx.Rebuild()
	if len(first) != len(second) {
		t.Fatalf("expected identical length")
	}
	for i := range first {
		if first[i].ID != second[i].ID {
			t.Fatalf("rebuild order changed")
		}
	}
}

func TestQueryFiltersAndMigration(t *testing.T) {
	idx := NewIndex()
	idx.Insert(Document{ID: "d1", Scope: "scope-S7-feed", Body: "query match", From: "user:alice", HasFile: true, HasLink: true, CreatedAt: time.Date(2025, 12, 1, 10, 0, 0, 0, time.UTC)})
	idx.Insert(Document{ID: "d2", Scope: "scope-other", Body: "query skip", From: "user:bob", HasFile: false, HasLink: true, CreatedAt: time.Date(2025, 11, 1, 10, 0, 0, 0, time.UTC)})
	filters := QueryFilters{
		FromUser: "user:alice",
		Range: [2]time.Time{
			time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 12, 2, 0, 0, 0, 0, time.UTC),
		},
		HasFile: true,
		HasLink: true,
	}
	results, err := idx.Query("scope-S7-feed", "query", filters, Pagination{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(results) != 1 || results[0].ID != "d1" {
		t.Fatalf("unexpected query results: %v", results)
	}
	if results[0].Scope != "scope-S7-feed" {
		t.Fatalf("unexpected scope %s", results[0].Scope)
	}
	mig := idx.ApplyMigration("v1.0")
	if mig.ID != "v1.0" {
		t.Fatalf("unexpected migration id")
	}
	if idx.MigrationStatus() != "applied" {
		t.Fatalf("expected applied status")
	}
}

func TestIndexScopeEnforcement(t *testing.T) {
	idx := NewIndex()
	idx.Insert(Document{ID: "scope-doc", Scope: "scope-S7-feed", Body: "scope match", CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)})
	if count := idx.ScopeInfo("scope-S7-feed"); count != 1 {
		t.Fatalf("expected one document for scope, got %d", count)
	}
	if count := idx.ScopeInfo("scope-other"); count != 0 {
		t.Fatalf("expected zero documents for other scope, got %d", count)
	}
	results, err := idx.Query("scope-other", "scope", QueryFilters{}, Pagination{Limit: 5})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no results for wrong scope, got %d", len(results))
	}
}

func TestIndexFilterCombinations(t *testing.T) {
	idx := NewIndex()
	idx.Insert(Document{ID: "f1", Scope: "scope-S7-feed", HasFile: true, HasLink: true, Body: "match", CreatedAt: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)})
	idx.Insert(Document{ID: "f2", Scope: "scope-S7-feed", HasFile: true, HasLink: false, Body: "match", CreatedAt: time.Date(2024, 12, 2, 0, 0, 0, 0, time.UTC)})
	filters := QueryFilters{HasFile: true, HasLink: true}
	results, err := idx.Query("scope-S7-feed", "", filters, Pagination{Limit: 10})
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(results) != 1 || results[0].ID != "f1" {
		t.Fatalf("unexpected filter results: %v", results)
	}
}

func TestEnsureScopeAndValidateFilters(t *testing.T) {
	if err := EnsureScope(""); err == nil {
		t.Fatalf("expected scope error")
	}
	if err := ValidateFilters(QueryFilters{Range: [2]time.Time{time.Time{}, time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}}); err == nil {
		t.Fatalf("expected invalid date range error")
	}
}
