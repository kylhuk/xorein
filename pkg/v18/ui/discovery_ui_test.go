package ui

import (
	"strings"
	"testing"

	"github.com/aether/code_aether/pkg/v18/directory"
	"github.com/aether/code_aether/pkg/v18/discoveryclient"
)

func TestRendererSummarizesEntries(t *testing.T) {
	entries := []directory.SignedEntry{
		directory.NewSignedEntry(directory.DirectoryEntry{NodeID: "node-1", Relay: "relay:alpha", Endpoint: "https://alpha", LastSeen: 1}, "signer"),
	}
	warnings := []discoveryclient.TrustWarning{{NodeID: "node-1", Message: "warn"}}
	report := Renderer{}.Summarize(entries, warnings)
	if !strings.Contains(report, "node-1") {
		t.Fatalf("expected node entry in summary")
	}
	if !strings.Contains(report, "warn") {
		t.Fatalf("expected warning message")
	}
}
