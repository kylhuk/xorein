package security

import (
	"fmt"
	"strings"
)

// ControlSeverity annotates how serious a failing control is.
type ControlSeverity string

const (
	SeverityCritical ControlSeverity = "critical"
	SeverityHigh     ControlSeverity = "high"
	SeverityMedium   ControlSeverity = "medium"
	SeverityLow      ControlSeverity = "low"
)

// HardeningCheck is a deterministic representation of a single
// security hardening validation step.
type HardeningCheck struct {
	Name     string
	Passed   bool
	Severity ControlSeverity
	Evidence string
}

// HardeningProfile contains the list of controls for a release gate.
type HardeningProfile struct {
	Name     string
	Controls []HardeningCheck
}

// ComplianceScore returns the percentage of passing controls.
func (p HardeningProfile) ComplianceScore() int {
	if len(p.Controls) == 0 {
		return 100
	}
	var passed int
	for _, c := range p.Controls {
		if c.Passed {
			passed++
		}
	}
	return (passed * 100) / len(p.Controls)
}

// CriticalFailures lists all failing controls marked as critical or high.
func (p HardeningProfile) CriticalFailures() []HardeningCheck {
	var failures []HardeningCheck
	for _, c := range p.Controls {
		if !c.Passed && (c.Severity == SeverityCritical || c.Severity == SeverityHigh) {
			failures = append(failures, c)
		}
	}
	return failures
}

// Summary provides deterministic reporting text based on profile state.
func (p HardeningProfile) Summary() string {
	if len(p.Controls) == 0 {
		return "no-controls"
	}
	var builder strings.Builder
	fmt.Fprintf(&builder, "profile=%s score=%d failures=%d", p.Name, p.ComplianceScore(), len(p.CriticalFailures()))
	return builder.String()
}

// Evaluate produces a deterministic result for downstream gating.
func Evaluate(profile HardeningProfile) HardeningResult {
	failures := profile.CriticalFailures()
	pass := len(failures) == 0
	return HardeningResult{
		Profile:  profile.Name,
		Score:    profile.ComplianceScore(),
		Pass:     pass,
		Message:  profile.Summary(),
		Failures: failures,
	}
}

// HardeningResult is the contract that release automation reads.
type HardeningResult struct {
	Profile  string
	Score    int
	Pass     bool
	Message  string
	Failures []HardeningCheck
}
