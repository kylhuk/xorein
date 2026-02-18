package v19

import (
	"testing"

	"github.com/aether/code_aether/pkg/v19/co"
)

func TestMobilityStepsTunnelBeforeRelay(t *testing.T) {
	ladder := co.NewPathLadder()
	current := co.PathTypeDirectQUIC
	steps := []struct {
		blocked bool
		want    co.PathType
	}{
		{blocked: true, want: co.PathTypeDirectTCP},
		{blocked: true, want: co.PathTypeTunnel},
		{blocked: true, want: co.PathTypeRelay},
	}

	for _, step := range steps {
		current, _, _ = ladder.Next(current, step.blocked, co.ReasonRecovery)
		if current != step.want {
			t.Fatalf("expected mobility ladder to reach %s, got %s", step.want, current)
		}
	}
}
