package conformance

import "testing"

func TestGateChecklistCoverage(t *testing.T) {
	check := NewGateChecklist(GateV7G1)
	check.RecordEvidence("S7-02", "doc referred")
	if check.IsSatisfied() {
		t.Fatalf("expected gate to be unsatisfied when evidence missing")
	}
	if missing := check.MissingScopes(); len(missing) == 0 {
		t.Fatalf("expected missing scopes for incomplete gate")
	}
	check.RecordEvidence("S7-03", "filled")
	check.RecordEvidence("S7-04", "filled")
	check.Completed = true
	if !check.IsSatisfied() {
		t.Fatalf("expected gate to be satisfied after completing evidence")
	}
	if GateCoverageScore(check) != 1 {
		t.Fatalf("expected coverage score 1, got %f", GateCoverageScore(check))
	}
}

func TestGateCoverageScoreIgnoresNonRequiredScopes(t *testing.T) {
	check := NewGateChecklist(GateV7G0)
	check.RecordEvidence("S7-01", "required")
	check.RecordEvidence("S7-99", "non-required")

	if score := GateCoverageScore(check); score != 0.5 {
		t.Fatalf("expected non-required evidence to be ignored, got %f", score)
	}
}

func TestGateIsComplete(t *testing.T) {
	check := NewGateChecklist(GateV7G0)
	check.Completed = true
	check.RecordEvidence("S7-01", "done")
	check.RecordEvidence("S7-02", "done")
	if !GateIsComplete(check) {
		t.Fatalf("expected gate to be complete when evidence present")
	}
}
