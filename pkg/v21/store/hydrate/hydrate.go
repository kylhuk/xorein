package hydrate

import (
	"strconv"

	"github.com/aether/code_aether/pkg/v21/store"
)

// Page describes a hydration page with cursors and state.
type Page struct {
	Events []store.TimelineEvent
	Next   store.HydrationCursor
	Prev   store.HydrationCursor
	State  store.EmptyHistoryState
}

// Hydrator provides deterministic ordering, pagination, and state metadata.
type Hydrator struct {
	events  []store.TimelineEvent
	cleared bool
	missing bool
}

// SetEvents replaces the stored timeline events and resets state flags.
func (h *Hydrator) SetEvents(events []store.TimelineEvent) {
	h.events = store.OrderedEvents(events)
	h.cleared = false
	h.missing = false
}

// MarkCleared indicates the timeline was cleared locally.
func (h *Hydrator) MarkCleared() {
	h.events = nil
	h.cleared = true
	h.missing = false
}

// MarkMissingBackfill flags that history is intentionally missing (no backfill).
func (h *Hydrator) MarkMissingBackfill() {
	if len(h.events) == 0 {
		h.missing = true
	}
}

// State returns the empty-history state that should be surfaced when no events exist.
func (h *Hydrator) State() store.EmptyHistoryState {
	if len(h.events) == 0 {
		if h.cleared {
			return store.EmptyHistoryCleared
		}
		if h.missing {
			return store.EmptyHistoryMissing
		}
		return store.EmptyHistoryNoData
	}
	return store.EmptyHistoryNone
}

// Paginate returns a deterministic page of events starting at the provided cursor index.
func (h *Hydrator) Paginate(cursor store.HydrationCursor, limit int) Page {
	if limit <= 0 {
		limit = 10
	}
	start := 0
	if cursor != "" {
		if parsed, err := strconv.Atoi(string(cursor)); err == nil && parsed >= 0 && parsed < len(h.events) {
			start = parsed
		}
	}

	end := start + limit
	if end > len(h.events) {
		end = len(h.events)
	}

	page := Page{
		Events: append([]store.TimelineEvent(nil), h.events[start:end]...),
		State:  h.State(),
	}

	if end < len(h.events) {
		page.Next = store.CursorFromIndex(end)
	}
	if start > 0 {
		prevIndex := start - limit
		if prevIndex < 0 {
			prevIndex = 0
		}
		page.Prev = store.CursorFromIndex(prevIndex)
	}

	return page
}
