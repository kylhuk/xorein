package store

import (
	"fmt"
	"sort"
	"time"
)

// StoreFailureReason is a deterministic failure reason emitted by the v21 store.
type StoreFailureReason string

const (
	StoreLocked            StoreFailureReason = "STORE_LOCKED"
	StoreCorrupt           StoreFailureReason = "STORE_CORRUPT"
	StoreMigrationRequired StoreFailureReason = "STORE_MIGRATION_REQUIRED"
	StoreQuotaExceeded     StoreFailureReason = "STORE_QUOTA_EXCEEDED"
)

// CanonicalEvent identifies when two inserts refer to the same timeline event.
type CanonicalEvent struct {
	MessageID string
	EventID   string
	ChannelID string
	SenderID  string
}

// Key returns a deterministic string representation of the canonical identity.
func (ce CanonicalEvent) Key() string {
	return fmt.Sprintf("%s:%s:%s:%s", ce.ChannelID, ce.MessageID, ce.EventID, ce.SenderID)
}

// TimelineEvent is the store-side representation of a chat event.
type TimelineEvent struct {
	Canonical CanonicalEvent
	Timestamp time.Time
	Payload   string
	ChannelID string
	Pinned    bool
	System    bool
}

// HydrationCursor encodes a cursor between pagination pages.
type HydrationCursor string

// CursorFromIndex produces a cursor that points to the provided index.
func CursorFromIndex(index int) HydrationCursor {
	return HydrationCursor(fmt.Sprintf("%d", index))
}

// EmptyHistoryState describes the namespace of empty timeline states exposed to clients.
type EmptyHistoryState string

const (
	EmptyHistoryNone    EmptyHistoryState = ""
	EmptyHistoryNoData  EmptyHistoryState = "no local history yet"
	EmptyHistoryCleared EmptyHistoryState = "history cleared locally"
	EmptyHistoryMissing EmptyHistoryState = "history missing (backfill not supported in v21)"
)

// OrderedEvents returns a copy of events sorted by timestamp (oldest first), resolving ties by canonical key.
func OrderedEvents(events []TimelineEvent) []TimelineEvent {
	sorted := append([]TimelineEvent(nil), events...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Timestamp.Equal(sorted[j].Timestamp) {
			return sorted[i].Canonical.Key() < sorted[j].Canonical.Key()
		}
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})
	return sorted
}
