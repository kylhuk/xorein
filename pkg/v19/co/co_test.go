package co

import "testing"

func TestPathLadder_Transitions(t *testing.T) {
	ladder := NewPathLadder()
	current := PathTypeDirectQUIC
	next, reason, changed := ladder.Next(current, true, ReasonStartup)
	if !changed || next != PathTypeDirectTCP || reason != ReasonStartup {
		t.Fatalf("expected transition to tcp on startup failure, got %v %v %v", next, reason, changed)
	}

	next, reason, changed = ladder.Next(PathTypeDirectTCP, true, ReasonMessaging)
	if next != PathTypeTunnel || reason != ReasonTunnelHeal || !changed {
		t.Fatalf("expected tunnel transition with tunnel-heal reason, got %v %v %v", next, reason, changed)
	}

	// No failure should keep same path.
	current = PathTypeTunnel
	next, reason, changed = ladder.Next(current, false, ReasonRecovery)
	if next != current || reason != ReasonRecovery || changed {
		t.Fatalf("expected no transition on success, got %v %v %v", next, reason, changed)
	}
}

func TestPathLadder_FinalStepNoTransition(t *testing.T) {
	ladder := NewPathLadder()
	current := PathTypeTURN
	next, reason, changed := ladder.Next(current, true, ReasonCallHandoff)
	if next != current || reason != ReasonCallHandoff || changed {
		t.Fatalf("final step should stay same but still include context reason, got %v %v %v", next, reason, changed)
	}
}
