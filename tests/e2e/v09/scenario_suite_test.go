package v09e2e

import (
	"strings"
	"testing"

	"github.com/aether/code_aether/pkg/v09/conformance"
	"github.com/aether/code_aether/pkg/v09/relay"
	"github.com/aether/code_aether/pkg/v09/scale"
	"github.com/aether/code_aether/pkg/v09/scenario"
)

func TestV09ScenarioSuite(t *testing.T) {
	t.Run("E2E-V9-FORGE", func(t *testing.T) {
		if err := scenario.RunForgeScenario(); err != nil {
			t.Fatalf("forge scenario failed: %v", err)
		}
	})

	t.Run("E2E-V9-SCALE-FANOUT", func(t *testing.T) {
		plan := scale.HierarchyPlan(2_048)
		if len(plan) == 0 {
			t.Fatalf("hierarchy plan empty")
		}
		if plan[0].Fanout != 64 {
			t.Fatalf("edge fanout = %d, want 64", plan[0].Fanout)
		}
	})

	t.Run("E2E-V9-RELAY-RECOVERY", func(t *testing.T) {
		posture := relay.RecoveryPlan(96)
		if !strings.Contains(posture.NextStep, "throttle") {
			t.Fatalf("expected throttling recovery at high utilization, got %s", posture.NextStep)
		}
	})

	t.Run("E2E-V9-CONFORMANCE-GATE", func(t *testing.T) {
		checklist := conformance.Checklist{"VA-I1": true}
		result, err := conformance.ValidateChecklist("V9-G1", checklist)
		if err != nil {
			t.Fatalf("unexpected error %v", err)
		}
		if result.Ready {
			t.Fatalf("expected gate not ready when artifacts missing")
		}
		if len(result.Missing) == 0 {
			t.Fatalf("expected missing artifacts, got none")
		}
	})
}
