package v19

import (
	"testing"

	"github.com/aether/code_aether/pkg/v11/relaypolicy"
	"github.com/aether/code_aether/pkg/v19/co"
)

func TestNatMatrixDeterministicFallback(t *testing.T) {
	ladder := co.NewPathLadder()
	scenarios := []struct {
		name       string
		current    co.PathType
		context    co.Reason
		want       co.PathType
		wantReason co.Reason
	}{
		{name: "full-cone", current: co.PathTypeDirectQUIC, context: co.ReasonStartup, want: co.PathTypeDirectTCP, wantReason: co.ReasonStartup},
		{name: "restricted", current: co.PathTypeDirectTCP, context: co.ReasonMessaging, want: co.PathTypeTunnel, wantReason: co.ReasonTunnelHeal},
		{name: "symmetric", current: co.PathTypeTunnel, context: co.ReasonCallHandoff, want: co.PathTypeRelay, wantReason: co.ReasonCallHandoff},
		{name: "tcp-only", current: co.PathTypeRelay, context: co.ReasonRecovery, want: co.PathTypeTURN, wantReason: co.ReasonRecovery},
	}

	for _, scenario := range scenarios {
		next, reason, changed := ladder.Next(scenario.current, true, scenario.context)
		if !changed || next != scenario.want || reason != scenario.wantReason {
			t.Fatalf("nat %s expected %s:%s got %s:%s", scenario.name, scenario.want, scenario.wantReason, next, reason)
		}
	}

	next, reason, changed := ladder.Next(co.PathTypeRelay, false, co.ReasonMessaging)
	if changed || next != co.PathTypeRelay || reason != co.ReasonMessaging {
		t.Fatalf("expected deterministic stay on relay, got %s %s", next, reason)
	}
}

func TestNATMatrixIncludesRelayBoundary(t *testing.T) {
	if err := relaypolicy.ValidateMode(relaypolicy.PersistenceModeSessionMetadata); err != nil {
		t.Fatalf("expected allowed relay mode, got %v", err)
	}

	err := relaypolicy.ValidateMode(relaypolicy.PersistenceModeDurableMessageBody)
	if err == nil {
		t.Fatal("expected durable relay mode to be rejected")
	}
}
