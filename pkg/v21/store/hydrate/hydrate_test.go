package hydrate

import (
	"reflect"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v21/store"
)

func timelineEvent(id string, chron time.Time) store.TimelineEvent {
	return store.TimelineEvent{
		Canonical: store.CanonicalEvent{MessageID: id, EventID: id + "-v1", ChannelID: "chan", SenderID: "sender"},
		Timestamp: chron,
		Payload:   id,
		ChannelID: "chan",
	}
}

func TestOrderingAndPagination(t *testing.T) {
	events := []store.TimelineEvent{
		timelineEvent("c", time.Unix(3, 0)),
		timelineEvent("a", time.Unix(1, 0)),
		timelineEvent("b", time.Unix(2, 0)),
	}

	var h Hydrator
	h.SetEvents(events)

	first := h.Paginate("", 2)
	if len(first.Events) != 2 || first.Events[0].Payload != "a" || first.Events[1].Payload != "b" {
		t.Fatalf("unexpected first page events: %#v", first.Events)
	}
	if first.Next == "" {
		t.Fatalf("expected next cursor, got empty")
	}

	second := h.Paginate(first.Next, 2)
	if len(second.Events) != 1 || second.Events[0].Payload != "c" {
		t.Fatalf("unexpected second page: %#v", second.Events)
	}
}

func TestEmptyHistoryStates(t *testing.T) {
	var h Hydrator
	if state := h.State(); state != store.EmptyHistoryNoData {
		t.Fatalf("expected no-data state, got %s", state)
	}

	h.MarkMissingBackfill()
	if state := h.State(); state != store.EmptyHistoryMissing {
		t.Fatalf("expected missing state, got %s", state)
	}

	h.MarkCleared()
	if state := h.State(); state != store.EmptyHistoryCleared {
		t.Fatalf("expected cleared state, got %s", state)
	}
}

func TestPaginateCursorStability(t *testing.T) {
	events := []store.TimelineEvent{
		timelineEvent("d", time.Unix(4, 0)),
		timelineEvent("a", time.Unix(1, 0)),
		timelineEvent("b", time.Unix(2, 0)),
		timelineEvent("c", time.Unix(3, 0)),
	}
	var h Hydrator
	h.SetEvents(events)

	first := h.Paginate("", 2)
	second := h.Paginate(first.Next, 2)
	repeatSecond := h.Paginate(first.Next, 2)
	if !reflect.DeepEqual(payloads(second.Events), payloads(repeatSecond.Events)) {
		t.Fatalf("second page not stable: %v vs %v", payloads(second.Events), payloads(repeatSecond.Events))
	}
	if second.Prev == "" {
		t.Fatalf("expected prev cursor from second page")
	}
	prev := h.Paginate(second.Prev, 2)
	if !reflect.DeepEqual(payloads(prev.Events), payloads(first.Events)) {
		t.Fatalf("prev page mismatch: %v vs %v", payloads(prev.Events), payloads(first.Events))
	}
	if prev.Next != first.Next {
		t.Fatalf("expected prev page next to match original next, got %s vs %s", prev.Next, first.Next)
	}
	// repeated calls to prev should also be stable
	if !reflect.DeepEqual(payloads(h.Paginate(second.Prev, 2).Events), payloads(first.Events)) {
		t.Fatalf("repeated prev call unstable")
	}
}

func payloads(events []store.TimelineEvent) []string {
	result := make([]string, len(events))
	for i, event := range events {
		result[i] = event.Payload
	}
	return result
}
