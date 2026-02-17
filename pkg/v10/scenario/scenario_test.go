package scenario

import (
	"testing"

	"github.com/aether/code_aether/pkg/v10/security"
)

func TestRunGenesisScenario(t *testing.T) {
	t.Parallel()

	if err := RunGenesisScenario(); err != nil {
		t.Fatalf("unexpected failure running genesis scenario: %v", err)
	}

	t.Run("security mappings", func(t *testing.T) {
		if len(security.AssetInventory()) < 3 {
			t.Fatal("asset inventory missing entries")
		}
		if len(security.ThreatModel()) != 3 {
			t.Fatalf("unexpected threat model size: %d", len(security.ThreatModel()))
		}
	})
}
