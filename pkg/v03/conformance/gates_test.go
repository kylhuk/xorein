package conformance

import "testing"

func TestGateChecklistSatisfaction(t *testing.T) {
	scope := []string{"S3-01", "S3-02"}
	checklist := NewGateChecklist("GATE", scope)

	checklist.RecordEvidence("S3-01", "doc-1")
	checklist.RecordEvidence("S3-02", "doc-2")
	checklist.Completed = true

	if !checklist.IsSatisfied() {
		t.Fatalf("checklist should be satisfied with all evidence: %+v", checklist)
	}
}

func TestGateChecklistTraceSummary(t *testing.T) {
	checklist := NewGateChecklist("G1", []string{"S3-01"})
	summary := checklist.TraceSummary()
	if summary == "" {
		t.Fatal("expected non-empty trace summary")
	}
}

func TestEvidenceMatrixCopies(t *testing.T) {
	positive := []string{"p"}
	matrix := EvidenceMatrix(positive, nil, nil, nil)
	if matrix["positive"][0] != "p" {
		t.Fatal("positive evidence missing")
	}
	positive[0] = "mutated"
	if matrix["positive"][0] != "p" {
		t.Fatal("matrix should copy slices")
	}
}

func TestTraceableBulletsCount(t *testing.T) {
	if len(TraceableBullets) != 17 {
		t.Fatalf("expected 17 traceable bullets, got %d", len(TraceableBullets))
	}
}
