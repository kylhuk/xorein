package reputation

import (
	"fmt"
	"math"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

type ReputationReasonClass string

const (
	ReputationReasonWeight ReputationReasonClass = "VA-R1:reputation.weight"
)

type ReputationContract struct {
	ScopeID  string
	TaskID   string
	Artifact string
	Gate     conformance.GateID
	Reason   ReputationReasonClass
}

func NewReputationContract(scope, task, artifact string, gate conformance.GateID, reason ReputationReasonClass) ReputationContract {
	return ReputationContract{
		ScopeID:  scope,
		TaskID:   task,
		Artifact: artifact,
		Gate:     gate,
		Reason:   reason,
	}
}

func (c ReputationContract) EvidenceAnchor() string {
	return fmt.Sprintf("%s#%s", c.Artifact, c.TaskID)
}

func (c ReputationContract) ReasonLabel() string {
	return string(c.Reason)
}

func ReputationReasonClasses() []ReputationReasonClass {
	return []ReputationReasonClass{ReputationReasonWeight}
}

type ReputationWeight struct {
	Weight     float64
	Confidence float64
	Reason     string
}

func ComputeReputationWeight(raw float64, trustPaths int, uncertainty float64) ReputationWeight {
	bounded := math.Max(0, math.Min(raw, 1))
	confidence := math.Min(1, float64(trustPaths)/5)
	reason := "reputation.weight.success"
	if trustPaths < 2 || uncertainty > 0.5 {
		reason = "reputation.weight.failure"
		bounded *= 0.8
	}
	bounded = math.Max(0, bounded)
	return ReputationWeight{Weight: bounded, Confidence: confidence, Reason: reason}
}
