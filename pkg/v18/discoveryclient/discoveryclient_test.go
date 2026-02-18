package discoveryclient

import (
	"testing"

	"github.com/aether/code_aether/pkg/v18/directory"
	"github.com/aether/code_aether/pkg/v18/indexer"
)

func TestMergeResponsesDetectsTrustWarnings(t *testing.T) {
	idxA := indexer.NewIndexer(18)
	idxB := indexer.NewIndexer(18)
	entry := directory.DirectoryEntry{NodeID: "node-1", Endpoint: "endpoint", LastSeen: 1}
	idxA.Add(directory.NewSignedEntry(entry, "signer-a"))
	idxB.Add(directory.NewSignedEntry(entry, "signer-b"))

	client := NewClient()
	result := client.MergeResponses([]indexer.SignedResponse{idxA.SignedResponse(), idxB.SignedResponse()})
	if len(result.Entries) != 1 {
		t.Fatalf("expected one merged entry, got %d", len(result.Entries))
	}
	if len(result.Warnings) == 0 {
		t.Fatalf("expected warning on signature mismatch")
	}
}

func TestJoinFunnelStateRecordsAttempts(t *testing.T) {
	client := NewClient()
	client.RecordStage(JoinStageHandshake, false)
	client.RecordStage(JoinStageCompleted, true)
	if !client.Funnel.Failed {
		t.Fatalf("expected failed flag after handshake failure")
	}
	if client.Funnel.Attempts != 1 {
		t.Fatalf("expected 1 completed attempt, got %d", client.Funnel.Attempts)
	}
	summary := client.FunnelSummary()
	if summary == "" {
		t.Fatalf("expected non-empty funnel summary")
	}
}
