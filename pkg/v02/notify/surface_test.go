package notify

import "testing"

func TestResolveSurfaceRule(t *testing.T) {
	tests := []struct {
		name          string
		contextType   ContextType
		wantPlace     BadgePlacement
		wantHighlight HighlightMode
	}{
		{name: "dm uses inbox and message highlight", contextType: ContextTypeDM, wantPlace: BadgePlacementInbox, wantHighlight: HighlightMessage},
		{name: "group dm uses thread and message highlight", contextType: ContextTypeGroupDM, wantPlace: BadgePlacementThread, wantHighlight: HighlightMessage},
		{name: "server uses channel and channel highlight", contextType: ContextTypeServer, wantPlace: BadgePlacementChannel, wantHighlight: HighlightChannel},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rule := ResolveSurfaceRule(tc.contextType)
			if rule.Placement != tc.wantPlace || rule.Highlight != tc.wantHighlight || !rule.ClearOnRead {
				t.Fatalf("ResolveSurfaceRule(%q)=%+v", tc.contextType, rule)
			}
		})
	}
}

func TestApplyAttentionEvent(t *testing.T) {
	base := SurfaceState{
		Counter:     CounterState{Counts: map[string]uint32{}, Seen: map[string]struct{}{}},
		Highlights:  map[string]bool{},
		ContextType: map[string]ContextType{},
	}

	inactiveMention := ApplyAttentionEvent(base, AttentionEvent{ContextID: "dm-1", ContextType: ContextTypeDM, DedupeKey: "evt-1", Mention: true})
	if inactiveMention.BadgeCount != 1 || !inactiveMention.Highlight || inactiveMention.Reason != SurfaceReasonHighlighted {
		t.Fatalf("inactive mention mismatch: %+v", inactiveMention)
	}

	activeState := inactiveMention.Next
	activeState.Counter.ActiveContextID = "dm-1"
	activeMention := ApplyAttentionEvent(activeState, AttentionEvent{ContextID: "dm-1", ContextType: ContextTypeDM, DedupeKey: "evt-2", Mention: true})
	if activeMention.BadgeCount != 1 || activeMention.Highlight || activeMention.Reason != SurfaceReasonActiveContext {
		t.Fatalf("active mention mismatch: %+v", activeMention)
	}

	duplicate := ApplyAttentionEvent(inactiveMention.Next, AttentionEvent{ContextID: "dm-1", ContextType: ContextTypeDM, DedupeKey: "evt-1", Mention: true})
	if duplicate.BadgeCount != 1 || duplicate.Reason != SurfaceReasonDuplicateIgnored {
		t.Fatalf("duplicate mismatch: %+v", duplicate)
	}
}

func TestSurfaceClearOnRead(t *testing.T) {
	state := SurfaceState{
		Counter: CounterState{
			Counts: map[string]uint32{"chan-1": 4},
			Seen:   map[string]struct{}{},
		},
		Highlights:  map[string]bool{"chan-1": true},
		ContextType: map[string]ContextType{"chan-1": ContextTypeServer},
	}

	opened := OpenAttentionContext(state, "chan-1")
	if opened.BadgeCount != 0 || opened.Highlight || opened.Reason != SurfaceReasonReadCleared || opened.Next.Counter.ActiveContextID != "chan-1" {
		t.Fatalf("OpenAttentionContext mismatch: %+v", opened)
	}

	marked := MarkAttentionRead(opened.Next, "chan-1")
	if marked.BadgeCount != 0 || marked.Highlight || marked.Reason != SurfaceReasonReadCleared {
		t.Fatalf("MarkAttentionRead mismatch: %+v", marked)
	}
}
