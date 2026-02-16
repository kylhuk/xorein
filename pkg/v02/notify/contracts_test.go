package notify

import "testing"

func TestApplyReceive(t *testing.T) {
	tests := []struct {
		name       string
		state      CounterState
		event      Event
		wantCount  uint32
		wantReason CounterReason
	}{
		{
			name: "increments when context is inactive",
			state: CounterState{
				ActiveContextID: "ctx-b",
				Counts:          map[string]uint32{},
				Seen:            map[string]struct{}{},
			},
			event:      Event{ContextID: "ctx-a", DedupeKey: "evt-1"},
			wantCount:  1,
			wantReason: ReasonIncremented,
		},
		{
			name: "suppresses increment in active context",
			state: CounterState{
				ActiveContextID: "ctx-a",
				Counts:          map[string]uint32{},
				Seen:            map[string]struct{}{},
			},
			event:      Event{ContextID: "ctx-a", DedupeKey: "evt-1"},
			wantCount:  0,
			wantReason: ReasonActiveContextSuppressed,
		},
		{
			name: "ignores duplicate events by dedupe key",
			state: CounterState{
				ActiveContextID: "ctx-b",
				Counts:          map[string]uint32{"ctx-a": 1},
				Seen:            map[string]struct{}{"evt-1": {}},
			},
			event:      Event{ContextID: "ctx-a", DedupeKey: "evt-1"},
			wantCount:  1,
			wantReason: ReasonDuplicateIgnored,
		},
		{
			name: "invalid event does not mutate counters",
			state: CounterState{
				ActiveContextID: "ctx-a",
				Counts:          map[string]uint32{"ctx-a": 2},
				Seen:            map[string]struct{}{},
			},
			event:      Event{ContextID: "", DedupeKey: "evt-2"},
			wantCount:  0,
			wantReason: ReasonInvalidEvent,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ApplyReceive(tc.state, tc.event)
			if got.Reason != tc.wantReason || got.Count != tc.wantCount {
				t.Fatalf("ApplyReceive(%+v,%+v)=(count=%d,reason=%q) want (count=%d,reason=%q)", tc.state, tc.event, got.Count, got.Reason, tc.wantCount, tc.wantReason)
			}
		})
	}
}

func TestOpenContext(t *testing.T) {
	state := CounterState{
		ActiveContextID: "ctx-b",
		Counts:          map[string]uint32{"ctx-a": 3},
		Seen:            map[string]struct{}{},
	}
	got := OpenContext(state, "ctx-a")
	if got.Reason != ReasonContextOpenReset || got.Next.ActiveContextID != "ctx-a" || got.Next.Counts["ctx-a"] != 0 {
		t.Fatalf("OpenContext mismatch: %+v", got)
	}
}

func TestMarkRead(t *testing.T) {
	state := CounterState{
		Counts: map[string]uint32{"ctx-a": 5},
		Seen:   map[string]struct{}{},
	}
	got := MarkRead(state, "ctx-a")
	if got.Reason != ReasonMarkedRead || got.Next.Counts["ctx-a"] != 0 {
		t.Fatalf("MarkRead mismatch: %+v", got)
	}
}

func TestLeaveContext(t *testing.T) {
	state := CounterState{
		ActiveContextID: "ctx-a",
		Counts:          map[string]uint32{"ctx-a": 2},
		Seen:            map[string]struct{}{},
	}
	left := LeaveContext(state, "ctx-a")
	if left.Reason != ReasonContextLeft || left.Next.ActiveContextID != "" {
		t.Fatalf("LeaveContext active mismatch: %+v", left)
	}
	noop := LeaveContext(state, "ctx-b")
	if noop.Reason != ReasonNoop || noop.Next.ActiveContextID != "ctx-a" {
		t.Fatalf("LeaveContext noop mismatch: %+v", noop)
	}
}
