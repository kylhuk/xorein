package search

import (
	"strings"
	"testing"
	"time"
)

func TestNormalizeQuery(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	filters := QueryFilters{FromUser: "user:alpha", Range: [2]time.Time{now.Add(-time.Hour), now}, HasFile: true}
	got, err := NormalizeQuery("hello", filters)
	if err != nil {
		t.Fatalf("expected query normalization success, got err: %v", err)
	}
	if !strings.Contains(got, "from:user:alpha") {
		t.Fatalf("expected from filter, got %s", got)
	}
	if !strings.Contains(got, "has:file") {
		t.Fatalf("expected has:file hint")
	}
}

func TestNormalizeQueryInvalidRange(t *testing.T) {
	now := time.Unix(1700000000, 0).UTC()
	tests := []QueryFilters{
		{Range: [2]time.Time{now, time.Time{}}},
		{Range: [2]time.Time{time.Time{}, now}},
		{Range: [2]time.Time{now, now.Add(-time.Hour)}},
	}

	for _, filters := range tests {
		if _, err := NormalizeQuery("hello", filters); err != ErrInvalidDateRange {
			t.Fatalf("expected ErrInvalidDateRange, got %v", err)
		}
	}
}

func TestNormalizePagination(t *testing.T) {
	page := NormalizePagination(0, -1)
	if page.Limit != 10 || page.Offset != 0 {
		t.Fatalf("expected defaults limit=10 offset=0, got %v", page)
	}
	pageLarge := NormalizePagination(999, 5)
	if pageLarge.Limit != 100 {
		t.Fatalf("expected limit cap 100, got %d", pageLarge.Limit)
	}
}

func TestScopeAuthorized(t *testing.T) {
	if !ScopeAuthorized("S7-03", "user:beta") {
		t.Fatalf("expected authorized for user")
	}
	if ScopeAuthorized("X1", "bot:gamma") {
		t.Fatalf("expected unauthorized for bot")
	}
}
