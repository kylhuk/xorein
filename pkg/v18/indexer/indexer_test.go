package indexer

import (
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v18/directory"
)

func TestIndexerProducesDeterministicResponse(t *testing.T) {
	originalNow := now
	now = func() time.Time { return time.Unix(142, 0) }
	defer func() { now = originalNow }()

	idx := NewIndexer(18)
	idx.Add(directory.NewSignedEntry(directory.DirectoryEntry{NodeID: "node-z", Endpoint: "z", LastSeen: 1}, "signer"))
	idx.Add(directory.NewSignedEntry(directory.DirectoryEntry{NodeID: "node-a", Endpoint: "a", LastSeen: 2}, "signer"))

	resp := idx.SignedResponse()
	if resp.Timestamp != 142 {
		t.Fatalf("expected timestamp 142, got %d", resp.Timestamp)
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp.Entries))
	}
	if resp.Entries[0].Entry.NodeID != "node-a" {
		t.Fatalf("entries not sorted: %v", resp.Entries)
	}
	expected := signResponse(resp.Entries, resp.Version)
	if resp.Signature != expected {
		t.Fatalf("signature mismatch, expected %s got %s", expected, resp.Signature)
	}
}
