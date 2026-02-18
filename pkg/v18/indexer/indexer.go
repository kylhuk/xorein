package indexer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aether/code_aether/pkg/v18/directory"
)

var now = time.Now

// Indexer collects signed directory entries and exposes deterministic responses.
type Indexer struct {
	entries map[string]directory.SignedEntry
	version int
}

// NewIndexer creates an indexer for a specific slice version.
func NewIndexer(version int) *Indexer {
	if version == 0 {
		version = 18
	}
	return &Indexer{entries: make(map[string]directory.SignedEntry), version: version}
}

// Add ingests or updates a signed entry.
func (i *Indexer) Add(entry directory.SignedEntry) {
	existing, ok := i.entries[entry.Entry.NodeID]
	if !ok || entry.Entry.LastSeen >= existing.Entry.LastSeen {
		i.entries[entry.Entry.NodeID] = entry
	}
}

// SignedResponse is the deterministically ordered payload returned by the indexer.
type SignedResponse struct {
	Version   int
	Timestamp int64
	Entries   []directory.SignedEntry
	Signature string
}

// SignedResponse returns the snapshot of the current directory view.
func (i *Indexer) SignedResponse() SignedResponse {
	entries := make([]directory.SignedEntry, 0, len(i.entries))
	for _, entry := range i.entries {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(a, b int) bool {
		return entries[a].Entry.NodeID < entries[b].Entry.NodeID
	})
	return SignedResponse{
		Version:   i.version,
		Timestamp: now().Unix(),
		Entries:   entries,
		Signature: signResponse(entries, i.version),
	}
}

func signResponse(entries []directory.SignedEntry, version int) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("version=%d;", version))
	for _, entry := range entries {
		builder.WriteString(entry.Entry.NodeID)
		builder.WriteByte(':')
		builder.WriteString(entry.Entry.Endpoint)
		builder.WriteByte(';')
	}
	sum := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(sum[:])
}
