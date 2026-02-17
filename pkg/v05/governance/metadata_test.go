package governance

import (
	"testing"

	"github.com/aether/code_aether/pkg/v05/conformance"
)

func TestOpenDecisionsPreserved(t *testing.T) {
	if len(OpenDecisions) != 5 {
		t.Fatalf("expected five open decision IDs, got %d", len(OpenDecisions))
	}
	if OpenDecisions[0] != OpenDecisionOD501 {
		t.Fatalf("unexpected first decision ID: %s", OpenDecisions[0])
	}
}

func TestReleaseChecklistWorkflow(t *testing.T) {
	checklist := NewReleaseChecklist()
	for _, item := range ReleaseChecklistItems {
		checklist.Mark(item)
	}

	if !checklist.IsComplete() {
		t.Fatal("checklist should be complete after marking every item")
	}
}

func TestReleaseMetadataRecords(t *testing.T) {
	meta := NewReleaseMetadata()
	decision := OpenDecisionRecord{ID: OpenDecisionOD503, Gate: conformance.GateV5G5, Status: "Open"}
	meta.AddDecision(decision)

	if len(meta.Decisions) != 1 {
		t.Fatalf("expected one decision, got %d", len(meta.Decisions))
	}

	meta.RecordGateEvidence(conformance.GateV5G1, "evidence details")
	if detail := meta.GateEvidence[conformance.GateV5G1]; detail != "evidence details" {
		t.Fatalf("unexpected gate evidence: %s", detail)
	}
}
