package ui

import "fmt"

// EnforcementMode surfaces official-client enforcement intent.
type EnforcementMode string

const (
	EnforcementModeRelaxed EnforcementMode = "relaxed"
	EnforcementModeStrict  EnforcementMode = "strict"
)

// StatusSignal represents the UI contract for enforcement state.
type StatusSignal struct {
	Mode         EnforcementMode
	Sequence     int
	Reason       string
	TrustWarning string
}

// NewStatusSignal builds a deterministic UI payload.
func NewStatusSignal(mode EnforcementMode, sequence int, reason string, trusted bool) StatusSignal {
	warning := ""
	if !trusted {
		warning = "unverified enforcement"
	}
	return StatusSignal{Mode: mode, Sequence: sequence, Reason: reason, TrustWarning: warning}
}

// Summary returns the canonical string clients observe.
func (s StatusSignal) Summary() string {
	if s.TrustWarning != "" {
		return fmt.Sprintf("[%s/%d] %s (%s)", s.Mode, s.Sequence, s.Reason, s.TrustWarning)
	}
	return fmt.Sprintf("[%s/%d] %s", s.Mode, s.Sequence, s.Reason)
}

// IsStrict returns whether the enforcement mode is hard.
func (s StatusSignal) IsStrict() bool {
	return s.Mode == EnforcementModeStrict
}
