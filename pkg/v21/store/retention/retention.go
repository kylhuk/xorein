package retention

import (
	"sort"
	"time"

	"github.com/aether/code_aether/pkg/v21/store"
)

// RetentionPolicy captures defaults used by Phase 1 retention.
type RetentionPolicy struct {
	WindowDays int
	MaxEntries int
}

const clearHistoryReason = "CLEAR_LOCAL_HISTORY"

// ClearLocalHistoryResult describes the contractual reset state.
type ClearLocalHistoryResult struct {
	TimelineReset       bool
	SearchCoverageReset bool
	Reason              string
}

// PruneResult exposes the kept and evicted events.
type PruneResult struct {
	Kept   []store.TimelineEvent
	Pruned []store.TimelineEvent
}

// NewRetentionPolicy returns the default Phase 1 retention policy.
func NewRetentionPolicy() *RetentionPolicy {
	return &RetentionPolicy{
		WindowDays: 30,
		MaxEntries: 1000,
	}
}

// Prune evicts the oldest non-pinned, non-system events per channel while enforcing window and max entries.
func (rp *RetentionPolicy) Prune(events []store.TimelineEvent, now time.Time) PruneResult {
	if rp == nil {
		rp = NewRetentionPolicy()
	}

	windowStart := now.AddDate(0, 0, -rp.WindowDays)
	channelBuckets := map[string][]store.TimelineEvent{}
	for _, event := range events {
		channelBuckets[event.ChannelID] = append(channelBuckets[event.ChannelID], event)
	}

	kept := make([]store.TimelineEvent, 0, len(events))
	pruned := make([]store.TimelineEvent, 0, len(events))
	keptPerChannel := map[string]int{}

	channels := make([]string, 0, len(channelBuckets))
	for channel := range channelBuckets {
		channels = append(channels, channel)
	}
	sort.Strings(channels)
	for _, channel := range channels {
		bucket := channelBuckets[channel]
		ordered := store.OrderedEvents(bucket)
		for _, event := range ordered {
			if event.Pinned || event.System {
				kept = append(kept, event)
				continue
			}
			if !event.Timestamp.IsZero() && event.Timestamp.Before(windowStart) {
				pruned = append(pruned, event)
				continue
			}
			if rp.MaxEntries > 0 && keptPerChannel[channel] >= rp.MaxEntries {
				pruned = append(pruned, event)
				continue
			}
			kept = append(kept, event)
			keptPerChannel[channel]++
		}
	}

	return PruneResult{Kept: kept, Pruned: pruned}
}

// ClearLocalHistory resets the timeline/search coverage state contractually.
func (rp *RetentionPolicy) ClearLocalHistory() ClearLocalHistoryResult {
	return ClearLocalHistoryResult{
		TimelineReset:       true,
		SearchCoverageReset: true,
		Reason:              clearHistoryReason,
	}
}
