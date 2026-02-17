package governance

import (
	"testing"

	"github.com/aether/code_aether/pkg/v07/conformance"
)

func TestReleaseChecklistComplete(t *testing.T) {
	check := NewReleaseChecklist()
	for item := range check.Items {
		check.Mark(item)
	}
	if !check.IsComplete() {
		t.Fatalf("expected checklist to be complete after marking all items")
	}
}

func TestGateEvidenceRecording(t *testing.T) {
	meta := NewReleaseMetadata()
	meta.RecordGateEvidence(conformance.GateV7G0, "doc ready")
	if meta.GateCoverage() != 1 {
		t.Fatalf("unexpected gate coverage count: %d", meta.GateCoverage())
	}
	meta.AddDecision(OpenDecisionRecord{ID: OpenDecisionOD701, Gate: conformance.GateV7G1, Status: "Open", Notes: "needs more research"})
	if len(meta.Decisions) != 1 {
		t.Fatalf("expected one decision recorded")
	}
}
