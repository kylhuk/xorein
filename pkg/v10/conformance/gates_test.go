package conformance

import "testing"

func TestGateCollectionsAreDeterministic(t *testing.T) {
	t.Parallel()

	catalog := Catalog()
	if len(catalog) < 5 {
		t.Fatalf("expected at least five gates, got %d", len(catalog))
	}

	result, err := ValidateChecklist("V10-G0", Checklist{"VA-G1": true, "VA-G2": true, "VA-G3": true, "VA-G4": true, "VA-G5": true, "VA-G6": true})
	if err != nil {
		t.Fatalf("unexpected error validating checklist: %v", err)
	}
	if !result.Ready {
		t.Fatalf("expected V10-G0 to be ready when all artifacts present, got missing %v", result.Missing)
	}

	_, err = ValidateChecklist("unknown", Checklist{})
	if err == nil {
		t.Fatalf("expected unknown gate to return error")
	}
}

func TestSummariesReportEachGate(t *testing.T) {
	t.Parallel()

	fullChecklist := Checklist{}
	for _, gate := range Catalog() {
		for _, artifact := range gate.RequiredArtifacts {
			fullChecklist[artifact] = true
		}
	}

	results := Summaries(fullChecklist)
	if len(results) != len(Catalog()) {
		t.Fatalf("summaries count mismatch: got %d, want %d", len(results), len(Catalog()))
	}

	text := SummaryText(results)
	if text == "" {
		t.Fatal("expected summary text to describe gates")
	}
}
