package hardening

import (
	"fmt"
)

// ContinuityMode describes release expectations for runtime continuity.
type ContinuityMode string

const (
	ContinuitySteady    ContinuityMode = "steady"
	ContinuityResilient ContinuityMode = "resilient"
)

// ContinuityPolicy defines acceptable restart/downtime thresholds.
type ContinuityPolicy struct {
	Mode           ContinuityMode
	MaxRestarts    int
	MaxDowntimeSec int
}

// DefaultPolicy returns deterministic thresholds for each mode.
func DefaultPolicy(mode ContinuityMode) ContinuityPolicy {
	switch mode {
	case ContinuityResilient:
		return ContinuityPolicy{Mode: mode, MaxRestarts: 3, MaxDowntimeSec: 120}
	default:
		return ContinuityPolicy{Mode: ContinuitySteady, MaxRestarts: 1, MaxDowntimeSec: 30}
	}
}

// ContinuityObservation describes a runtime incident measured during release verification.
type ContinuityObservation struct {
	Description string
	Restarts    int
	DowntimeSec int
}

// Validate ensures an observation meets policy.
func (p ContinuityPolicy) Validate(obs ContinuityObservation) error {
	if obs.Restarts > p.MaxRestarts {
		return fmt.Errorf("restarts %d exceed allowed %d", obs.Restarts, p.MaxRestarts)
	}
	if obs.DowntimeSec > p.MaxDowntimeSec {
		return fmt.Errorf("downtime %ds exceeds allowed %ds", obs.DowntimeSec, p.MaxDowntimeSec)
	}
	return nil
}

// ContinuityReport aggregates observations after policy validation.
type ContinuityReport struct {
	Policy       ContinuityPolicy
	Observations []ContinuityObservation
}

// Issues returns deterministic list of validation failures.
func (r ContinuityReport) Issues() []string {
	var issues []string
	for _, obs := range r.Observations {
		if err := r.Policy.Validate(obs); err != nil {
			issues = append(issues, fmt.Sprintf("%s: %v", obs.Description, err))
		}
	}
	return issues
}

// IsCompliant reports whether all observations meet the policy.
func (r ContinuityReport) IsCompliant() bool {
	return len(r.Issues()) == 0
}
