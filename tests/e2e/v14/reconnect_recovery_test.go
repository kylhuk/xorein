package v14

import (
	"testing"

	"github.com/aether/code_aether/pkg/v14/ui"
	"github.com/aether/code_aether/pkg/v14/voice"
)

func TestReconnectRecovery(t *testing.T) {
	session := voice.NewSession()
	if _, _, err := session.Negotiate([]string{"opus/16000/2"}, []string{"tcp"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	first := session.Reconnect()
	second := session.Reconnect()
	if !(second > first) {
		t.Fatalf("expected increasing backoff, got %v <= %v", second, first)
	}

	state := ui.CallState{QualityScore: 40, RecoveryHint: true}
	if msg := ui.NoLimboMessage(state.QualityScore, state.RecoveryHint); msg != "Reconnecting with recovery-first flow..." {
		t.Fatalf("unexpected recovery message: %s", msg)
	}
}
