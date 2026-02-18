package v17_test

import (
	"testing"

	"github.com/aether/code_aether/pkg/v11/relaypolicy"
	"github.com/aether/code_aether/pkg/v17/moderation"
)

func TestAdversarialModerationEvents(t *testing.T) {
	engine := moderation.NewEngine()
	good := moderation.SignedEvent{ID: "evt-adversarial-1", Room: "room:gamma", Actor: "mod@relay", Target: "user:z", Type: moderation.EventTimeout, Timestamp: 300, Signature: "sig:mod@relay"}
	if res := engine.Apply(good); !res.Accepted {
		t.Fatalf("expected acceptance, reason %s", res.Reason)
	}
	forged := good
	forged.Signature = "sig:unknown"
	forged.ID = "evt-adversarial-2"
	if res := engine.Apply(forged); res.Reason != moderation.ReasonInvalidSignature {
		t.Fatalf("expected invalid signature, got %s", res.Reason)
	}
}

func TestRelayNoDataRegression(t *testing.T) {
	if err := relaypolicy.ValidateMode(relaypolicy.PersistenceModeDurableMessageBody); err == nil {
		t.Fatal("expected durable mode to be rejected")
	}
}
