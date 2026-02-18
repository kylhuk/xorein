package co

// PathType describes the transport layer used for a leg of the connectivity orchestrator.
type PathType string

const (
	PathTypeDirectQUIC PathType = "direct-quic"
	PathTypeDirectTCP  PathType = "direct-tcp"
	PathTypeTunnel     PathType = "tunnel"
	PathTypeRelay      PathType = "relay"
	PathTypeTURN       PathType = "turn"
)

// pathOrder defines the deterministic ladder used for failover.
var pathOrder = []PathType{
	PathTypeDirectQUIC,
	PathTypeDirectTCP,
	PathTypeTunnel,
	PathTypeRelay,
	PathTypeTURN,
}

// Reason describes why a path transition occurred.
type Reason string

const (
	ReasonStartup     Reason = "startup"
	ReasonMessaging   Reason = "messaging"
	ReasonCallHandoff Reason = "call-handoff"
	ReasonRecovery    Reason = "recovery"
	ReasonProbeTimed  Reason = "probe-timed"
	ReasonTunnelHeal  Reason = "tunnel-heal"
)

// PathLadder deterministically selects the next path type when a failure is reported.
type PathLadder struct {
	steps []PathType
}

// NewPathLadder allocates a ladder that follows the defined order.
func NewPathLadder() PathLadder {
	steps := make([]PathType, len(pathOrder))
	copy(steps, pathOrder)
	return PathLadder{steps: steps}
}

// Next selects the next path type given the active path and whether the last attempt has failed.
// It returns the chosen path, reason, and a flag indicating whether a transition happened.
func (l PathLadder) Next(current PathType, failed bool, context Reason) (PathType, Reason, bool) {
	if !failed {
		return current, context, false
	}

	idx := indexOf(current, l.steps)
	if idx < 0 || idx == len(l.steps)-1 {
		return current, reasonForTransition(current, current, context), false
	}

	next := l.steps[idx+1]
	return next, reasonForTransition(current, next, context), true
}

func indexOf(value PathType, values []PathType) int {
	for i, v := range values {
		if v == value {
			return i
		}
	}
	return -1
}

func reasonForTransition(from, to PathType, context Reason) Reason {
	if to == PathTypeTunnel && from != PathTypeTunnel {
		return ReasonTunnelHeal
	}

	switch context {
	case ReasonStartup:
		return ReasonStartup
	case ReasonCallHandoff:
		return ReasonCallHandoff
	case ReasonMessaging:
		return ReasonMessaging
	case ReasonRecovery:
		return ReasonRecovery
	default:
		return ReasonProbeTimed
	}
}
