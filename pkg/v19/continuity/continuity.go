package continuity

// State describes the continuity contract for reading and calls.
type State string

const (
	StateDraft State = "draft"
	StateRead  State = "read"
	StateCall  State = "call"
	StateWake  State = "wake"
)

// Event describes deterministic triggers.
type Event string

const (
	EventDraftPersist Event = "draft-persist"
	EventReadPersist  Event = "read-persist"
	EventCallHandoff  Event = "call-handoff"
	EventWakeSignal   Event = "wake-signal"
)

// Transition returns the next continuity state and a reason for the change.
func Transition(current State, event Event) (State, string) {
	switch event {
	case EventDraftPersist:
		return StateDraft, "draft-saved"
	case EventReadPersist:
		return StateRead, "read-synced"
	case EventCallHandoff:
		return StateCall, "handoff-complete"
	case EventWakeSignal:
		return StateWake, "wake-ack"
	default:
		return current, "noop"
	}
}
