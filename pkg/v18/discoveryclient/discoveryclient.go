package discoveryclient

import (
	"fmt"
	"sort"

	"github.com/aether/code_aether/pkg/v18/directory"
	"github.com/aether/code_aether/pkg/v18/indexer"
)

// TrustWarning flags inconsistent descriptors during merge.
type TrustWarning struct {
	NodeID  string
	Message string
}

// MergeResult contains the deduplicated entry set and any warnings.
type MergeResult struct {
	Entries  []directory.SignedEntry
	Warnings []TrustWarning
}

// Client simulates a discovery client merging indexer responses.
type Client struct {
	Funnel JoinFunnelState
}

// NewClient initializes the discovery client.
func NewClient() *Client {
	return &Client{Funnel: JoinFunnelState{Stage: JoinStageDiscovery}}
}

// MergeResponses merges multiple indexer snapshots, deduplicating by NodeID.
func (c *Client) MergeResponses(responses []indexer.SignedResponse) MergeResult {
	nodeMap := make(map[string]directory.SignedEntry)
	sigMap := make(map[string]string)
	var warnings []TrustWarning
	for idx, resp := range responses {
		for _, entry := range resp.Entries {
			stored, seen := nodeMap[entry.Entry.NodeID]
			if !seen || entry.Entry.LastSeen > stored.Entry.LastSeen {
				nodeMap[entry.Entry.NodeID] = entry
			}
			if prev, ok := sigMap[entry.Entry.NodeID]; ok && prev != entry.Signature {
				warnings = append(warnings, TrustWarning{
					NodeID:  entry.Entry.NodeID,
					Message: fmt.Sprintf("response[%d] signature mismatch", idx),
				})
			}
			sigMap[entry.Entry.NodeID] = entry.Signature
		}
	}
	entries := make([]directory.SignedEntry, 0, len(nodeMap))
	for _, entry := range nodeMap {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Entry.NodeID < entries[j].Entry.NodeID
	})
	return MergeResult{Entries: entries, Warnings: warnings}
}

// JoinStage enumerates the join funnel levels.
type JoinStage string

const (
	JoinStageUnknown   JoinStage = "unknown"
	JoinStageDiscovery JoinStage = "discovery"
	JoinStageHandshake JoinStage = "relay-handshake"
	JoinStageCompleted JoinStage = "completed"
)

// JoinFunnelState represents the current join progress.
type JoinFunnelState struct {
	Stage    JoinStage
	Failed   bool
	Attempts int
}

// RecordStage registers a stage transition and success indicator.
func (c *Client) RecordStage(stage JoinStage, success bool) {
	c.Funnel.Stage = stage
	if !success {
		c.Funnel.Failed = true
	}
	if stage == JoinStageCompleted && success {
		c.Funnel.Attempts++
	}
}

// FunnelSummary returns a compact description of funnel state.
func (c *Client) FunnelSummary() string {
	status := "ok"
	if c.Funnel.Failed {
		status = "failed"
	}
	return fmt.Sprintf("stage=%s attempts=%d status=%s", c.Funnel.Stage, c.Funnel.Attempts, status)
}
