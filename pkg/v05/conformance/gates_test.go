package conformance

import "testing"

func TestGateChecklistSatisfactionAndEvidence(t *testing.T) {
	checklist := NewGateChecklist(GateV5G2)
	checklist.Completed = true
	checklist.RecordEvidence("S5-04", "doc-1")
	checklist.RecordEvidence("S5-05", "doc-2")

	if !checklist.IsSatisfied() {
		t.Fatal("Gate should be satisfied when completed and evidence present")
	}

	checklist.Completed = false
	if checklist.IsSatisfied() {
		t.Fatal("Gate should not be satisfied when not marked completed")
	}
}

func TestGateCoverageScore(t *testing.T) {
	checklist := NewGateChecklist(GateV5G3)
	checklist.RecordEvidence("S5-06", "outline")

	if score := GateCoverageScore(checklist); score != 1.0 {
		t.Fatalf("expected full coverage for single scope, got %f", score)
	}

	checklist.RecordEvidence("S5-06", "updated")
	if score := GateCoverageScore(checklist); score != 1.0 {
		t.Fatalf("score should remain 1.0 after evidence overwrite, got %f", score)
	}

	checklist = NewGateChecklist(GateV5G0)
	checklist.RecordEvidence("S5-01", "required")
	checklist.RecordEvidence("S5-99", "non-required")
	if score := GateCoverageScore(checklist); score != 0.5 {
		t.Fatalf("expected out-of-scope evidence to be ignored, got %f", score)
	}
}

func TestRequiredScopesCopyAndScopeMapCoverage(t *testing.T) {
	checklist := NewGateChecklist(GateV5G0)
	scopes := checklist.RequiredScopes()
	if len(scopes) != 2 {
		t.Fatalf("expected two scopes, got %d", len(scopes))
	}

	scopes[0] = "mutated"
	fresh := checklist.RequiredScopes()
	if fresh[0] == "mutated" {
		t.Fatal("RequiredScopes should not expose internal slice")
	}

	for gate, mapped := range GateScopeMapping {
		for _, scope := range mapped {
			found := false
			for _, known := range ScopeIDs {
				if scope == known {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("scope %s mapped from gate %s is not in ScopeIDs", scope, gate)
			}
		}
	}
}
