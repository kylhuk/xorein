package directory

import "testing"

func TestNewSignedEntryIncludesSignature(t *testing.T) {
	entry := DirectoryEntry{NodeID: "node-1", Relay: "relay:alpha", Endpoint: "https://node:44", LastSeen: 100}
	signed := NewSignedEntry(entry, "signer@relay")
	if signed.Signature == "" {
		t.Fatal("expected signature, got empty string")
	}
	if err := signed.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestSortedMaintainsDeterministicOrder(t *testing.T) {
	raw := []SignedEntry{
		NewSignedEntry(DirectoryEntry{NodeID: "node-b", Endpoint: "b"}, "signer"),
		NewSignedEntry(DirectoryEntry{NodeID: "node-a", Endpoint: "a"}, "signer"),
	}
	sorted := Sorted(raw)
	ids := NodeIDs(sorted)
	if len(ids) != 2 || ids[0] != "node-a" || ids[1] != "node-b" {
		t.Fatalf("unexpected sorted order: %v", ids)
	}
}
