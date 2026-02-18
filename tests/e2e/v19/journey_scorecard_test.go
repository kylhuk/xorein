package v19

import (
	"testing"

	"github.com/aether/code_aether/pkg/v11/relaypolicy"
	"github.com/aether/code_aether/pkg/v19/co"
)

func TestJourneyScorecardReasonTaxonomy(t *testing.T) {
	ladder := co.NewPathLadder()
	path, reason, changed := ladder.Next(co.PathTypeDirectQUIC, true, co.ReasonCallHandoff)
	if !changed || reason != co.ReasonCallHandoff {
		t.Fatalf("expected call-handoff reason on fail, got %s %t", reason, changed)
	}
	if path != co.PathTypeDirectTCP {
		t.Fatalf("call-handoff should prefer tcp next, got %s", path)
	}

	err := relaypolicy.ValidateMode(relaypolicy.PersistenceModeDurableMessageBody)
	if err == nil {
		t.Fatalf("expected relay durability restriction to remain in force")
	}
}
