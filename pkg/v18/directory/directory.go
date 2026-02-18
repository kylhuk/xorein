package directory

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// DirectoryEntry captures the lightweight discovery descriptor for a node.
type DirectoryEntry struct {
	NodeID   string
	Relay    string
	Endpoint string
	LastSeen int64
}

// SignedEntry pairs a DirectoryEntry with a deterministic signature.
type SignedEntry struct {
	Entry     DirectoryEntry
	Signature string
}

// NewSignedEntry produces a deterministically signed descriptor using a signer label.
func NewSignedEntry(entry DirectoryEntry, signer string) SignedEntry {
	return SignedEntry{Entry: entry, Signature: signatureFor(entry, signer)}
}

func signatureFor(entry DirectoryEntry, signer string) string {
	var builder strings.Builder
	builder.WriteString(entry.NodeID)
	builder.WriteByte('|')
	builder.WriteString(entry.Relay)
	builder.WriteByte('|')
	builder.WriteString(entry.Endpoint)
	builder.WriteByte('|')
	builder.WriteString(fmt.Sprintf("%d", entry.LastSeen))
	builder.WriteByte('|')
	builder.WriteString(signer)
	sum := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(sum[:])
}

// Validate ensures the signed descriptor contains the required fields.
func (s SignedEntry) Validate() error {
	if strings.TrimSpace(s.Entry.NodeID) == "" {
		return fmt.Errorf("missing node id")
	}
	if strings.TrimSpace(s.Signature) == "" {
		return fmt.Errorf("missing signature for %s", s.Entry.NodeID)
	}
	return nil
}

// Sorted returns a new slice of entries ordered by NodeID.
func Sorted(entries []SignedEntry) []SignedEntry {
	result := make([]SignedEntry, len(entries))
	copy(result, entries)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Entry.NodeID < result[j].Entry.NodeID
	})
	return result
}

// NodeIDs extracts the deterministic list of node identifiers.
func NodeIDs(entries []SignedEntry) []string {
	ids := make([]string, len(entries))
	for i, entry := range entries {
		ids[i] = entry.Entry.NodeID
	}
	return ids
}
