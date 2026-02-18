package ui

import (
	"fmt"
	"strings"

	"github.com/aether/code_aether/pkg/v18/directory"
	"github.com/aether/code_aether/pkg/v18/discoveryclient"
)

// Renderer exposes deterministic discovery views.
type Renderer struct{}

// Summarize builds a lightweight textual report of the directory state.
func (Renderer) Summarize(entries []directory.SignedEntry, warnings []discoveryclient.TrustWarning) string {
	var builder strings.Builder
	builder.WriteString("Discovery summary:\n")
	for _, entry := range entries {
		fmt.Fprintf(&builder, "- %s via %s (%s)\n", entry.Entry.NodeID, entry.Entry.Relay, entry.Entry.Endpoint)
	}
	if len(warnings) > 0 {
		builder.WriteString("Warnings:\n")
		for _, warning := range warnings {
			fmt.Fprintf(&builder, "* %s: %s\n", warning.NodeID, warning.Message)
		}
	}
	return builder.String()
}
