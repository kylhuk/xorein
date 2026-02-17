package v1_0_test

import (
	"testing"

	"github.com/aether/code_aether/pkg/v10/conformance"
	"github.com/aether/code_aether/pkg/v10/docs"
	"github.com/aether/code_aether/pkg/v10/release"
	"github.com/aether/code_aether/pkg/v10/scenario"
)

func TestRunGenesisScenario(t *testing.T) {
	t.Parallel()

	if err := scenario.RunGenesisScenario(); err != nil {
		t.Fatalf("scenario failed: %v", err)
	}
}

func TestGateAlignedInvariants(t *testing.T) {
	t.Parallel()

	result, err := conformance.ValidateChecklist("V10-G0", conformance.Checklist{
		"VA-G1": true,
		"VA-G2": true,
		"VA-G3": true,
		"VA-G4": true,
		"VA-G5": true,
		"VA-G6": true,
	})
	if err != nil {
		t.Fatalf("unexpected gate validation error: %v", err)
	}
	if !result.Ready {
		t.Fatalf("expected gate ready, missing %v", result.Missing)
	}

	if len(release.ManifestChecklist()) == 0 {
		t.Fatal("release manifest checklist should list artifacts")
	}
	if got := len(docs.UserGuideChecklist()); got != 3 {
		t.Fatalf("user guide checklist len = %d, want 3", got)
	}
}
