package modsync

import (
	"sort"

	"github.com/aether/code_aether/pkg/v17/moderation"
)

// Orderer deterministically merges replication batches.
type Orderer struct {
	seen map[string]struct{}
}

// NewOrderer builds a new deterministic merger.
func NewOrderer() *Orderer {
	return &Orderer{seen: make(map[string]struct{})}
}

// Merge merges replication batches while dropping duplicates and
// imposing timestamp/id ordering.
func (o *Orderer) Merge(batches ...[]moderation.SignedEvent) []moderation.SignedEvent {
	merged := make([]moderation.SignedEvent, 0)
	for _, batch := range batches {
		for _, evt := range batch {
			if _, ok := o.seen[evt.ID]; ok {
				continue
			}
			o.seen[evt.ID] = struct{}{}
			merged = append(merged, evt)
		}
	}
	sort.Slice(merged, func(i, j int) bool {
		if merged[i].Timestamp != merged[j].Timestamp {
			return merged[i].Timestamp < merged[j].Timestamp
		}
		return merged[i].ID < merged[j].ID
	})
	return merged
}

// Known returns true if the event was already seen by the orderer.
func (o *Orderer) Known(event moderation.SignedEvent) bool {
	_, ok := o.seen[event.ID]
	return ok
}
