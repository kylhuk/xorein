package voice

import (
	"errors"
	"strings"
	"testing"
)

func TestSelectNoiseReducer(t *testing.T) {
	t.Run("studio mode", func(t *testing.T) {
		decision := SelectNoiseReducer(true, nil)
		if decision.Selected != RNNoise {
			t.Fatalf("expected RNNoise, got %s", decision.Selected)
		}
		if decision.Reason != "stable environment" {
			t.Fatalf("unexpected reason %s", decision.Reason)
		}
	})

	t.Run("baseline fallback", func(t *testing.T) {
		decision := SelectNoiseReducer(false, nil)
		if decision.Selected != DTLN {
			t.Fatalf("expected DTLN, got %s", decision.Selected)
		}
		if decision.Reason != "baseline fallback" {
			t.Fatalf("unexpected reason %s", decision.Reason)
		}
	})

	t.Run("fallback error", func(t *testing.T) {
		err := errors.New("unstable network")
		decision := SelectNoiseReducer(false, err)
		if decision.Selected != DTLN {
			t.Fatalf("expected DTLN, got %s", decision.Selected)
		}
		if !strings.Contains(decision.Reason, "unstable network") {
			t.Fatalf("reason missing error detail: %s", decision.Reason)
		}
	})
}
