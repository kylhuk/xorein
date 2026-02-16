package notify

type CounterReason string

const (
	ReasonIncremented             CounterReason = "incremented"
	ReasonDuplicateIgnored        CounterReason = "duplicate-ignored"
	ReasonActiveContextSuppressed CounterReason = "active-context-suppressed"
	ReasonContextOpenReset        CounterReason = "context-open-reset"
	ReasonMarkedRead              CounterReason = "marked-read"
	ReasonContextLeft             CounterReason = "context-left"
	ReasonNoop                    CounterReason = "noop"
	ReasonInvalidEvent            CounterReason = "invalid-event"
)

type Event struct {
	ContextID string
	DedupeKey string
}

type CounterState struct {
	ActiveContextID string
	Counts          map[string]uint32
	Seen            map[string]struct{}
}

type CounterDecision struct {
	Count  uint32
	Reason CounterReason
	Next   CounterState
}

func ApplyReceive(state CounterState, event Event) CounterDecision {
	next := copyState(state)
	if event.ContextID == "" || event.DedupeKey == "" {
		return CounterDecision{Count: 0, Reason: ReasonInvalidEvent, Next: next}
	}
	if _, ok := next.Seen[event.DedupeKey]; ok {
		return CounterDecision{Count: next.Counts[event.ContextID], Reason: ReasonDuplicateIgnored, Next: next}
	}
	next.Seen[event.DedupeKey] = struct{}{}
	if event.ContextID == next.ActiveContextID {
		return CounterDecision{Count: next.Counts[event.ContextID], Reason: ReasonActiveContextSuppressed, Next: next}
	}
	next.Counts[event.ContextID]++
	return CounterDecision{Count: next.Counts[event.ContextID], Reason: ReasonIncremented, Next: next}
}

func OpenContext(state CounterState, contextID string) CounterDecision {
	next := copyState(state)
	if contextID == "" {
		return CounterDecision{Count: 0, Reason: ReasonNoop, Next: next}
	}
	next.ActiveContextID = contextID
	next.Counts[contextID] = 0
	return CounterDecision{Count: 0, Reason: ReasonContextOpenReset, Next: next}
}

func MarkRead(state CounterState, contextID string) CounterDecision {
	next := copyState(state)
	if contextID == "" {
		return CounterDecision{Count: 0, Reason: ReasonNoop, Next: next}
	}
	next.Counts[contextID] = 0
	return CounterDecision{Count: 0, Reason: ReasonMarkedRead, Next: next}
}

func LeaveContext(state CounterState, contextID string) CounterDecision {
	next := copyState(state)
	if contextID == "" || next.ActiveContextID != contextID {
		return CounterDecision{Count: next.Counts[contextID], Reason: ReasonNoop, Next: next}
	}
	next.ActiveContextID = ""
	return CounterDecision{Count: next.Counts[contextID], Reason: ReasonContextLeft, Next: next}
}

func copyState(state CounterState) CounterState {
	counts := make(map[string]uint32, len(state.Counts))
	for key, value := range state.Counts {
		counts[key] = value
	}
	seen := make(map[string]struct{}, len(state.Seen))
	for key := range state.Seen {
		seen[key] = struct{}{}
	}
	return CounterState{
		ActiveContextID: state.ActiveContextID,
		Counts:          counts,
		Seen:            seen,
	}
}
