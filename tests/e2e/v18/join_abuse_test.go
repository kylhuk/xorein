package v18_test

import (
	"testing"

	"github.com/aether/code_aether/pkg/v18/discoveryclient"
)

func TestJoinFunnelResistsAbuse(t *testing.T) {
	client := discoveryclient.NewClient()
	client.RecordStage(discoveryclient.JoinStageDiscovery, true)
	client.RecordStage(discoveryclient.JoinStageHandshake, false)
	client.RecordStage(discoveryclient.JoinStageCompleted, true)

	if !client.Funnel.Failed {
		t.Fatalf("expected failure flag after handshake abuse")
	}
	if client.Funnel.Attempts != 1 {
		t.Fatalf("expected 1 completed attempt, got %d", client.Funnel.Attempts)
	}
	if summary := client.FunnelSummary(); summary == "" {
		t.Fatalf("expected non-empty summary")
	}
}
