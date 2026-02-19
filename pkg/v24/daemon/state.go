package daemon

import "fmt"

// State represents a deterministic daemon lifecycle phase.
type State string

const (
	// StateStopped indicates the daemon is idle / not running.
	StateStopped State = "stopped"

	// StateStarting indicates the daemon is in the startup phase.
	StateStarting State = "starting"

	// StateRunning indicates the daemon is accepting connections.
	StateRunning State = "running"

	// StateStopping indicates the daemon is winding down.
	StateStopping State = "stopping"

	// StateFailed indicates the daemon encountered a failure and needs reset.
	StateFailed State = "failed"
)

// TransitionError describes why a state change is disallowed.
type TransitionError struct {
	From State
	To   State
}

func (t *TransitionError) Error() string {
	return fmt.Sprintf("state %s -> %s is disallowed", t.From, t.To)
}

// allowedTransitions defines the deterministic state graph.
var allowedTransitions = map[State][]State{
	StateStopped:  {StateStarting, StateFailed},
	StateStarting: {StateRunning, StateFailed},
	StateRunning:  {StateStopping, StateFailed},
	StateStopping: {StateStopped, StateFailed},
	StateFailed:   {StateStopped},
}

// ValidateTransition returns nil if the change from `from` to `to` is allowed
// according to the deterministic lifecycle graph.
func ValidateTransition(from, to State) error {
	if from == to {
		return nil
	}

	if targets, ok := allowedTransitions[from]; ok {
		for _, candidate := range targets {
			if candidate == to {
				return nil
			}
		}
	}

	return &TransitionError{From: from, To: to}
}
