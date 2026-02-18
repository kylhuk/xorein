package v18perf_test

import (
	"fmt"
	"testing"

	"github.com/aether/code_aether/pkg/v18/directory"
	"github.com/aether/code_aether/pkg/v18/discoveryclient"
	"github.com/aether/code_aether/pkg/v18/indexer"
)

func TestDiscoveryStepsPerformance(t *testing.T) {
	idx := indexer.NewIndexer(18)
	for i := 0; i < 25; i++ {
		entry := directory.DirectoryEntry{
			NodeID:   fmt.Sprintf("node-%02d", i),
			Relay:    fmt.Sprintf("relay-%d", i%3),
			Endpoint: fmt.Sprintf("https://node-%02d", i),
			LastSeen: int64(i),
		}
		idx.Add(directory.NewSignedEntry(entry, fmt.Sprintf("perf-%d", i)))
	}
	resp := idx.SignedResponse()
	if len(resp.Entries) != 25 {
		t.Fatalf("expected 25 entries, got %d", len(resp.Entries))
	}
	modified := resp
	modified.Entries = make([]directory.SignedEntry, len(resp.Entries))
	copy(modified.Entries, resp.Entries)
	modified.Entries[0] = directory.NewSignedEntry(directory.DirectoryEntry{
		NodeID:   resp.Entries[0].Entry.NodeID,
		Relay:    resp.Entries[0].Entry.Relay,
		Endpoint: resp.Entries[0].Entry.Endpoint,
		LastSeen: resp.Entries[0].Entry.LastSeen + 5,
	}, "perf-warning")

	client := discoveryclient.NewClient()
	result := client.MergeResponses([]indexer.SignedResponse{resp, modified})
	if len(result.Entries) != 25 {
		t.Fatalf("expected 25 merged entries, got %d", len(result.Entries))
	}
	if len(result.Warnings) == 0 {
		t.Fatalf("expected trust warning from modified entry")
	}
}
