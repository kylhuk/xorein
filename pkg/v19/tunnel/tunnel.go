package tunnel

// Action describes the deterministic next step for a tunnel.
type Action string

const (
	ActionNoOp  Action = "noop"
	ActionRetry Action = "retry"
	ActionClose Action = "close"
)

// Decision is the output of the tunnel policy evaluation.
type Decision struct {
	Action Action
	Reason string
}

// Policy describes opportunistic tunnel establishment and healing limits.
type Policy struct {
	MaxRetries   int
	AutoTeardown bool
}

// DefaultPolicy is the deterministic policy used by v19.
func DefaultPolicy() Policy {
	return Policy{MaxRetries: 3, AutoTeardown: true}
}

// Evaluate returns the deterministic action for a tunnel based on the current state.
func (p Policy) Evaluate(attempts int, established bool, heartbeat bool) Decision {
	if established {
		if !heartbeat && attempts < p.MaxRetries {
			return Decision{Action: ActionRetry, Reason: "heartbeat-missed"}
		}
		if !heartbeat && p.AutoTeardown {
			return Decision{Action: ActionClose, Reason: "relay-timeout"}
		}
		return Decision{Action: ActionNoOp, Reason: "healthy"}
	}

	if attempts >= p.MaxRetries {
		return Decision{Action: ActionClose, Reason: "retry-limit"}
	}

	if heartbeat {
		return Decision{Action: ActionRetry, Reason: "establishing"}
	}

	return Decision{Action: ActionRetry, Reason: "probing"}
}
