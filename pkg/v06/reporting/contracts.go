package reporting

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

type ReportingReasonClass string

const (
	ReportingReasonRoute ReportingReasonClass = "VA-R2:reporting.route"
)

type ReportingContract struct {
	ScopeID  string
	TaskID   string
	Artifact string
	Gate     conformance.GateID
	Reason   ReportingReasonClass
}

func NewReportingContract(scope, task, artifact string, gate conformance.GateID, reason ReportingReasonClass) ReportingContract {
	return ReportingContract{
		ScopeID:  scope,
		TaskID:   task,
		Artifact: artifact,
		Gate:     gate,
		Reason:   reason,
	}
}

func (c ReportingContract) EvidenceAnchor() string {
	return fmt.Sprintf("%s#%s", c.Artifact, c.TaskID)
}

func (c ReportingContract) ReasonLabel() string {
	return string(c.Reason)
}

func ReportingReasonClasses() []ReportingReasonClass {
	return []ReportingReasonClass{ReportingReasonRoute}
}

type RoutingDecision struct {
	IdempotencyKey string
	Accept         bool
	Reason         string
	Retry          bool
	OrderingHint   int
}

func computeIdempotencyKey(reportID string) string {
	hash := sha256.Sum256([]byte(reportID))
	return hex.EncodeToString(hash[:])
}

func EvaluateReportRouting(reportID string, previousIDs []string, acked bool, orderingHint int, lastAck time.Time) RoutingDecision {
	sorted := make([]string, len(previousIDs))
	copy(sorted, previousIDs)
	sort.Strings(sorted)
	id := computeIdempotencyKey(reportID)
	decision := RoutingDecision{IdempotencyKey: id, OrderingHint: orderingHint}
	if acked && time.Since(lastAck) < time.Minute {
		decision.Accept = true
		decision.Reason = "reporting.route.accept"
		return decision
	}
	for _, previous := range sorted {
		if previous == reportID {
			decision.Accept = false
			decision.Retry = true
			decision.Reason = "reporting.route.failure"
			return decision
		}
	}
	decision.Accept = true
	decision.Reason = "reporting.route.accept"
	if orderingHint < len(sorted) {
		decision.Retry = true
		decision.Reason = "reporting.route.retry"
	}
	return decision
}
