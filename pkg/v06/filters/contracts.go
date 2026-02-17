package filters

import (
	"fmt"
	"time"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

type FilterReasonClass string

const (
	FilterReasonProcess FilterReasonClass = "VA-R3:filters.process"
)

type FilterContract struct {
	ScopeID  string
	TaskID   string
	Artifact string
	Gate     conformance.GateID
	Reason   FilterReasonClass
}

func NewFilterContract(scope, task, artifact string, gate conformance.GateID, reason FilterReasonClass) FilterContract {
	return FilterContract{
		ScopeID:  scope,
		TaskID:   task,
		Artifact: artifact,
		Gate:     gate,
		Reason:   reason,
	}
}

func (c FilterContract) EvidenceAnchor() string {
	return fmt.Sprintf("%s#%s", c.Artifact, c.TaskID)
}

func (c FilterContract) ReasonLabel() string {
	return string(c.Reason)
}

func FilterReasonClasses() []FilterReasonClass {
	return []FilterReasonClass{FilterReasonProcess}
}

type SecurityMode string

const (
	SecurityModeClear SecurityMode = "clear"
	SecurityModeE2EE  SecurityMode = "e2ee"
)

type FilterExecutionDecision struct {
	Mode     SecurityMode
	Allowed  bool
	Reason   string
	Recovery string
	Time     time.Time
}

func DecideFilterExecution(mode SecurityMode, optional bool) FilterExecutionDecision {
	decision := FilterExecutionDecision{Mode: mode, Time: time.Now()}
	switch mode {
	case SecurityModeClear:
		decision.Allowed = true
		decision.Reason = "filters.process.success"
	case SecurityModeE2EE:
		if optional {
			decision.Allowed = false
			decision.Reason = "filters.process.blocked"
			decision.Recovery = "filters.process.recover"
		} else {
			decision.Allowed = true
			decision.Reason = "filters.process.success"
		}
	default:
		decision.Allowed = optional
		decision.Reason = "filters.process.blocked"
	}
	return decision
}

type PolicyEnvelope struct {
	MinSeverity int
	MaxSeverity int
	Reason      string
}

type FilterPolicyDecision struct {
	Allowed      bool
	Severity     int
	Reason       string
	EnvelopeName string
}

func (p PolicyEnvelope) Evaluate(requestedSeverity int) FilterPolicyDecision {
	severity := requestedSeverity
	if severity < p.MinSeverity {
		severity = p.MinSeverity
	}
	if p.MaxSeverity > 0 && severity > p.MaxSeverity {
		severity = p.MaxSeverity
	}
	allowed := severity >= p.MinSeverity && (p.MaxSeverity == 0 || severity <= p.MaxSeverity)
	reason := "filters.policy.accept"
	if !allowed {
		reason = "filters.policy.blocked"
	}
	return FilterPolicyDecision{Allowed: allowed, Severity: severity, Reason: reason, EnvelopeName: p.Reason}
}
