package v14

import (
	"testing"

	"github.com/aether/code_aether/pkg/v14/voice"
)

func TestCallSetupSteps(t *testing.T) {
	session := voice.NewSession()
	ladder := session.Ladder()
	if len(ladder) != 4 {
		t.Fatalf("expected four fallback steps, got %d", len(ladder))
	}

	expected := []voice.FallbackStep{voice.FallbackDirect, voice.FallbackMesh, voice.FallbackSFU, voice.FallbackTURN}
	for i, step := range ladder {
		if step != expected[i] {
			t.Fatalf("fallback mismatch at %d: expected %q got %q", i, expected[i], step)
		}
	}

	session.AdvanceFallback()
	if session.CurrentFallback() != voice.FallbackMesh {
		t.Fatalf("expected mesh after advance, got %q", session.CurrentFallback())
	}
}
