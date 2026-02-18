package v14

import (
	"testing"

	"github.com/aether/code_aether/pkg/v11/relaypolicy"
	"github.com/aether/code_aether/pkg/v14/signaling"
	"github.com/aether/code_aether/pkg/v14/ui"
	"github.com/aether/code_aether/pkg/v14/voice"
)

func TestVoiceFlowSequence(t *testing.T) {
	lifecycle := signaling.NewLifecycle(2)
	if err := lifecycle.CreateRoom(); err != nil {
		t.Fatalf("create room failed: %v", err)
	}
	if err := lifecycle.JoinRoom(); err != nil {
		t.Fatalf("join room failed: %v", err)
	}

	session := voice.NewSession()
	if _, _, err := session.Negotiate([]string{"opus/48000/2"}, []string{"udp"}); err != nil {
		t.Fatalf("negotiation failed: %v", err)
	}

	if msg := ui.NoLimboMessage(90, false); msg == "" {
		t.Fatalf("expected stable message")
	}

	if err := relaypolicy.ValidateMode(relaypolicy.PersistenceModeDurableMessageBody); err == nil {
		t.Fatalf("relay policy should reject durable storage for voice")
	}

	allowed := relaypolicy.AllowedClasses()
	if len(allowed) == 0 {
		t.Fatalf("expected at least one allowed storage class")
	}
}
