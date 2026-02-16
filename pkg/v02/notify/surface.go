package notify

type ContextType string

const (
	ContextTypeDM      ContextType = "dm"
	ContextTypeGroupDM ContextType = "groupdm"
	ContextTypeServer  ContextType = "server"
)

type BadgePlacement string

const (
	BadgePlacementInbox   BadgePlacement = "inbox-list"
	BadgePlacementThread  BadgePlacement = "thread-list"
	BadgePlacementChannel BadgePlacement = "channel-list"
)

type HighlightMode string

const (
	HighlightNone    HighlightMode = "none"
	HighlightMessage HighlightMode = "message-highlight"
	HighlightChannel HighlightMode = "channel-highlight"
)

type SurfaceReason string

const (
	SurfaceReasonIncremented      SurfaceReason = "surface-incremented"
	SurfaceReasonHighlighted      SurfaceReason = "surface-highlighted"
	SurfaceReasonActiveContext    SurfaceReason = "surface-active-context"
	SurfaceReasonDuplicateIgnored SurfaceReason = "surface-duplicate-ignored"
	SurfaceReasonInvalidEvent     SurfaceReason = "surface-invalid-event"
	SurfaceReasonReadCleared      SurfaceReason = "surface-read-cleared"
	SurfaceReasonNoop             SurfaceReason = "surface-noop"
)

type SurfaceRule struct {
	Placement   BadgePlacement
	Highlight   HighlightMode
	ClearOnRead bool
}

type AttentionEvent struct {
	ContextID   string
	ContextType ContextType
	DedupeKey   string
	Mention     bool
}

type SurfaceState struct {
	Counter     CounterState
	Highlights  map[string]bool
	ContextType map[string]ContextType
}

type SurfaceDecision struct {
	BadgeCount uint32
	Highlight  bool
	Placement  BadgePlacement
	Reason     SurfaceReason
	Next       SurfaceState
}

func ResolveSurfaceRule(contextType ContextType) SurfaceRule {
	switch contextType {
	case ContextTypeDM:
		return SurfaceRule{Placement: BadgePlacementInbox, Highlight: HighlightMessage, ClearOnRead: true}
	case ContextTypeGroupDM:
		return SurfaceRule{Placement: BadgePlacementThread, Highlight: HighlightMessage, ClearOnRead: true}
	default:
		return SurfaceRule{Placement: BadgePlacementChannel, Highlight: HighlightChannel, ClearOnRead: true}
	}
}

func ApplyAttentionEvent(state SurfaceState, event AttentionEvent) SurfaceDecision {
	next := copySurfaceState(state)
	if event.ContextID == "" || event.DedupeKey == "" {
		rule := ResolveSurfaceRule(event.ContextType)
		return SurfaceDecision{BadgeCount: 0, Highlight: false, Placement: rule.Placement, Reason: SurfaceReasonInvalidEvent, Next: next}
	}
	contextType := event.ContextType
	if contextType == "" {
		contextType = next.ContextType[event.ContextID]
	}
	next.ContextType[event.ContextID] = contextType
	rule := ResolveSurfaceRule(contextType)

	counter := ApplyReceive(next.Counter, Event{ContextID: event.ContextID, DedupeKey: event.DedupeKey})
	next.Counter = counter.Next
	if event.ContextID == next.Counter.ActiveContextID {
		reason := mapCounterToSurfaceReason(counter.Reason)
		if reason == SurfaceReasonIncremented {
			reason = SurfaceReasonActiveContext
		}
		next.Highlights[event.ContextID] = false
		return SurfaceDecision{
			BadgeCount: next.Counter.Counts[event.ContextID],
			Highlight:  false,
			Placement:  rule.Placement,
			Reason:     reason,
			Next:       next,
		}
	}
	highlight := next.Highlights[event.ContextID]
	if event.Mention && rule.Highlight != HighlightNone && counter.Reason == ReasonIncremented {
		next.Highlights[event.ContextID] = true
		highlight = true
		return SurfaceDecision{
			BadgeCount: next.Counter.Counts[event.ContextID],
			Highlight:  highlight,
			Placement:  rule.Placement,
			Reason:     SurfaceReasonHighlighted,
			Next:       next,
		}
	}
	return SurfaceDecision{
		BadgeCount: next.Counter.Counts[event.ContextID],
		Highlight:  highlight,
		Placement:  rule.Placement,
		Reason:     mapCounterToSurfaceReason(counter.Reason),
		Next:       next,
	}
}

func OpenAttentionContext(state SurfaceState, contextID string) SurfaceDecision {
	next := copySurfaceState(state)
	rule := ResolveSurfaceRule(next.ContextType[contextID])
	if contextID == "" {
		return SurfaceDecision{BadgeCount: 0, Highlight: false, Placement: rule.Placement, Reason: SurfaceReasonNoop, Next: next}
	}
	decision := OpenContext(next.Counter, contextID)
	next.Counter = decision.Next
	next.Highlights[contextID] = false
	return SurfaceDecision{
		BadgeCount: 0,
		Highlight:  false,
		Placement:  rule.Placement,
		Reason:     SurfaceReasonReadCleared,
		Next:       next,
	}
}

func MarkAttentionRead(state SurfaceState, contextID string) SurfaceDecision {
	next := copySurfaceState(state)
	rule := ResolveSurfaceRule(next.ContextType[contextID])
	if contextID == "" {
		return SurfaceDecision{BadgeCount: 0, Highlight: false, Placement: rule.Placement, Reason: SurfaceReasonNoop, Next: next}
	}
	decision := MarkRead(next.Counter, contextID)
	next.Counter = decision.Next
	next.Highlights[contextID] = false
	return SurfaceDecision{
		BadgeCount: 0,
		Highlight:  false,
		Placement:  rule.Placement,
		Reason:     SurfaceReasonReadCleared,
		Next:       next,
	}
}

func LeaveAttentionContext(state SurfaceState, contextID string) SurfaceDecision {
	next := copySurfaceState(state)
	rule := ResolveSurfaceRule(next.ContextType[contextID])
	decision := LeaveContext(next.Counter, contextID)
	next.Counter = decision.Next
	return SurfaceDecision{
		BadgeCount: next.Counter.Counts[contextID],
		Highlight:  next.Highlights[contextID],
		Placement:  rule.Placement,
		Reason:     mapCounterToSurfaceReason(decision.Reason),
		Next:       next,
	}
}

func mapCounterToSurfaceReason(reason CounterReason) SurfaceReason {
	switch reason {
	case ReasonIncremented:
		return SurfaceReasonIncremented
	case ReasonDuplicateIgnored:
		return SurfaceReasonDuplicateIgnored
	case ReasonActiveContextSuppressed:
		return SurfaceReasonActiveContext
	case ReasonInvalidEvent:
		return SurfaceReasonInvalidEvent
	case ReasonContextOpenReset, ReasonMarkedRead:
		return SurfaceReasonReadCleared
	default:
		return SurfaceReasonNoop
	}
}

func copySurfaceState(state SurfaceState) SurfaceState {
	highlights := make(map[string]bool, len(state.Highlights))
	for contextID, value := range state.Highlights {
		highlights[contextID] = value
	}
	contextTypes := make(map[string]ContextType, len(state.ContextType))
	for contextID, value := range state.ContextType {
		contextTypes[contextID] = value
	}
	return SurfaceState{
		Counter:     copyState(state.Counter),
		Highlights:  highlights,
		ContextType: contextTypes,
	}
}
