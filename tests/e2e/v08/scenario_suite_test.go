package v08e2e

import (
	"testing"

	"github.com/aether/code_aether/pkg/v08/conformance"
	"github.com/aether/code_aether/pkg/v08/pinning"
	"github.com/aether/code_aether/pkg/v08/scenario"
	"github.com/aether/code_aether/pkg/v08/threads"
)

func TestV08ScenarioSuite(t *testing.T) {
	t.Run("E2E-V8-ECHO", func(t *testing.T) {
		if err := scenario.RunEchoContracts(); err != nil {
			t.Fatalf("echo scenario failed: %v", err)
		}
	})

	t.Run("E2E-V8-THREAD-LIFECYCLE", func(t *testing.T) {
		gates := conformance.Gates()
		if issues := conformance.ValidateChecklist(gates); len(issues) != 0 {
			t.Fatalf("gate checklist invalid: %v", issues)
		}
		trace := threads.ThreadTrace{ID: "v8-thread", CreatedDepth: 2, ReplyDepth: 3}
		if err := threads.ValidateReplyLineage(trace); err != nil {
			t.Fatalf("thread validation failed: %v", err)
		}
	})

	t.Run("E2E-V8-PINNING-ORDER", func(t *testing.T) {
		authorities := []pinning.PinAuthority{
			{ID: "alpha", Scope: pinning.ScopeTeam, Priority: 5},
			{ID: "beta", Scope: pinning.ScopePersonal, Priority: 1},
		}
		ordered := pinning.DeterministicOrder(authorities)
		if ordered[0].Scope != pinning.ScopePersonal {
			t.Fatalf("expected personal scope first, got %s", ordered[0].Scope)
		}
	})
}
