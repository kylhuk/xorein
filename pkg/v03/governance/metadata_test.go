package governance

import (
	"reflect"
	"testing"
)

func TestGateEvidenceNormalizedSorting(t *testing.T) {
	original := GateEvidence{Gate: GateV3G3, Artifacts: []string{"z", "a"}, Completed: true, Notes: "notes"}
	normalized := original.Normalized()

	if !reflect.DeepEqual(normalized.Artifacts, []string{"a", "z"}) {
		t.Fatalf("artifacts not sorted: %v", normalized.Artifacts)
	}
	if original.Artifacts[0] != "z" || original.Artifacts[1] != "a" {
		t.Fatal("original slice mutated")
	}
}

func TestGateEvidenceForFactory(t *testing.T) {
	evidence := GateEvidenceFor(GateV3G0, []string{"artifact"}, false, "note")
	if evidence.Gate != GateV3G0 || evidence.Completed != false || len(evidence.Artifacts) != 1 {
		t.Fatalf("unexpected evidence: %+v", evidence)
	}
}

func TestRequiredGatesOrder(t *testing.T) {
	if len(RequiredGates) != 7 {
		t.Fatalf("expected 7 required gates, got %d", len(RequiredGates))
	}
	if RequiredGates[0] != GateV3G0 || RequiredGates[6] != GateV3G6 {
		t.Fatalf("required gates ordering changed: %v", RequiredGates)
	}
}

func TestOpenDecisionReminder(t *testing.T) {
	reminder := OpenDecisionReminder{DecisionID: "dec-123", Status: "Open", Notes: "pending"}
	if !reminder.IsOpen() {
		t.Fatal("expected reminder to be open")
	}
	reminder.Status = "Closed"
	if reminder.IsOpen() {
		t.Fatal("closed reminder should not be open")
	}
}
