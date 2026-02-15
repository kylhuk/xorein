package phase7

import "testing"

func TestHistoryManagerApplyAndFetch(t *testing.T) {
	win := 4
	h := NewHistoryManager(win)
	sender := ParticipantID("sender1")
	entries := []HistoryEntry{{Sequence: 1}, {Sequence: 2}, {Sequence: 3}}
	if _, err := h.Apply(sender, entries); err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	more := []HistoryEntry{{Sequence: 2}, {Sequence: 4}}
	merged, err := h.Apply(sender, more)
	if err != nil {
		t.Fatalf("second apply failed: %v", err)
	}
	if len(merged) != 4 {
		t.Fatalf("expected 4 unique entries, got %d", len(merged))
	}
	if merged[0].Sequence != 1 || merged[len(merged)-1].Sequence != 4 {
		t.Fatalf("unexpected order %v", merged)
	}
	got, err := h.Fetch(sender, nil)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if len(got) != len(merged) {
		t.Fatalf("fetch returned %d entries, want %d", len(got), len(merged))
	}
}

func TestHistoryManagerWindowAndFallback(t *testing.T) {
	win := 2
	h := NewHistoryManager(win)
	sender := ParticipantID("peer")
	if _, err := h.Apply(sender, []HistoryEntry{{Sequence: 1}, {Sequence: 2}}); err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	entries, err := h.Fetch(sender, nil)
	if err != nil {
		t.Fatalf("fetch failed: %v", err)
	}
	if len(entries) != win {
		t.Fatalf("expected %d entries, got %d", win, len(entries))
	}
	fetched, err := h.Fetch(ParticipantID("missing"), func() ([]HistoryEntry, error) {
		return []HistoryEntry{{Sequence: 5}}, nil
	})
	if err != nil {
		t.Fatalf("fallback fetch failed: %v", err)
	}
	if len(fetched) != 1 || fetched[0].Sequence != 5 {
		t.Fatalf("unexpected fallback entries %v", fetched)
	}
}

func TestHistoryManagerWindowExceeded(t *testing.T) {
	h := NewHistoryManager(2)
	entries := make([]HistoryEntry, 3)
	for i := range entries {
		entries[i] = HistoryEntry{Sequence: uint64(i + 1)}
	}
	if _, err := h.Apply(ParticipantID("sender"), entries); err != ErrWindowExceeded {
		t.Fatalf("expected window exceeded, got %v", err)
	}
}
