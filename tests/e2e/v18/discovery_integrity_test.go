package v18_test

import (
	"testing"

	"github.com/aether/code_aether/pkg/v11/relaypolicy"
	"github.com/aether/code_aether/pkg/v18/directory"
	"github.com/aether/code_aether/pkg/v18/discoveryclient"
	"github.com/aether/code_aether/pkg/v18/indexer"
)

func TestDiscoveryIntegrityEnforcesRelayBoundary(t *testing.T) {
	idx := indexer.NewIndexer(18)
	nodes := []directory.DirectoryEntry{
		{NodeID: "node-1", Relay: "relay-alpha", Endpoint: "https://alpha", LastSeen: 1},
		{NodeID: "node-2", Relay: "relay-beta", Endpoint: "https://beta", LastSeen: 2},
	}
	for _, node := range nodes {
		idx.Add(directory.NewSignedEntry(node, "relay-idx"))
	}
	resp := idx.SignedResponse()

	client := discoveryclient.NewClient()
	merged := client.MergeResponses([]indexer.SignedResponse{resp})
	if len(merged.Entries) != len(nodes) {
		t.Fatalf("expected %d entries, got %d", len(nodes), len(merged.Entries))
	}
	if err := relaypolicy.ValidateMode(relaypolicy.PersistenceModeTransientMetadata); err != nil {
		t.Fatalf("relay policy regression: %v", err)
	}
	if forbidden := relaypolicy.ForbiddenClasses(); len(forbidden) == 0 {
		t.Fatalf("relay policy must forbid at least one persistence class")
	}
}
