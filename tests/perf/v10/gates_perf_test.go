package v10perf

import (
	"testing"

	"github.com/aether/code_aether/pkg/v10/conformance"
)

func TestAllGateThresholdsSatisfied(t *testing.T) {
	t.Parallel()

	checklist := conformance.Checklist{}
	for _, gate := range conformance.Catalog() {
		for _, artifact := range gate.RequiredArtifacts {
			checklist[artifact] = true
		}
	}

	for _, result := range conformance.Summaries(checklist) {
		if !result.Ready {
			t.Fatalf("gate %s should be ready with complete checklist", result.Gate.ID)
		}
	}
}

func TestGateThresholdDetectsMissingArtifact(t *testing.T) {
	t.Parallel()

	checklist := conformance.Checklist{"VA-G1": true}
	result, err := conformance.ValidateChecklist("V10-G0", checklist)
	if err != nil {
		t.Fatalf("unexpected error validating gate: %v", err)
	}
	if result.Ready {
		t.Fatal("expected gate to be marked not ready with missing artifacts")
	}
	if len(result.Missing) == 0 {
		t.Fatal("expected at least one missing artifact")
	}
}
