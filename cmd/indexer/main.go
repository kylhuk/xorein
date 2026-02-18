package main

import (
	"fmt"
	"os"

	"github.com/aether/code_aether/pkg/v18/directory"
	"github.com/aether/code_aether/pkg/v18/indexer"
)

func main() {
	idx := indexer.NewIndexer(18)
	samples := []directory.DirectoryEntry{
		{NodeID: "node-main-a", Relay: "relay:main", Endpoint: "https://main-a", LastSeen: 1},
		{NodeID: "node-main-b", Relay: "relay:main", Endpoint: "https://main-b", LastSeen: 2},
	}
	for _, entry := range samples {
		signed := directory.NewSignedEntry(entry, "cmd-indexer")
		if err := signed.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "invalid entry: %v\n", err)
			os.Exit(1)
		}
		idx.Add(signed)
	}
	resp := idx.SignedResponse()
	fmt.Fprintf(os.Stdout, "indexer v%d response: %d entries, signature=%s\n", resp.Version, len(resp.Entries), resp.Signature)
}
