package ui

// UIState describes canonical no-limbo UI states.
type UIState string

const (
	StateIdle       UIState = "idle"
	StateConnecting UIState = "connecting"
	StateInCall     UIState = "in-call"
	StateRecovering UIState = "recovering"
)

// Action describes the recovery-first guidance.
type Action string

const (
	ActionShowConnect  Action = "show-connect"
	ActionShowRecover  Action = "show-recover"
	ActionShowInCall   Action = "show-in-call"
	ActionNoVisualWork Action = "no-change"
)

// EvaluateState yields the deterministic next UI state and supporting action.
func EvaluateState(current UIState, healthy bool, wake bool) (UIState, Action) {
	if !healthy {
		if current == StateInCall {
			return StateRecovering, ActionShowRecover
		}
		return StateConnecting, ActionShowConnect
	}

	if wake {
		return StateInCall, ActionShowInCall
	}

	return current, ActionNoVisualWork
}
