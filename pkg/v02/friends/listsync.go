package friends

import (
	"sort"
	"strings"

	"github.com/aether/code_aether/pkg/v02/presence"
)

type ListTab string

const (
	ListTabOnline  ListTab = "online"
	ListTabOffline ListTab = "offline"
	ListTabPending ListTab = "pending"
)

type ProjectionReason string

const (
	ProjectionReasonAcceptedOnline  ProjectionReason = "accepted-online"
	ProjectionReasonAcceptedOffline ProjectionReason = "accepted-offline"
	ProjectionReasonPending         ProjectionReason = "pending"
	ProjectionReasonInvalidIdentity ProjectionReason = "invalid-identity"
)

type FriendListRecord struct {
	IdentityID           string
	RequestState         RequestState
	PresenceState        presence.State
	LastUpdatedUnix      int64
	PendingAction        bool
	SynchronizationError bool
}

type FriendListEntry struct {
	IdentityID    string
	Tab           ListTab
	RequestState  RequestState
	PresenceState presence.State
	Reason        ProjectionReason
}

type FriendsProjection struct {
	Online  []FriendListEntry
	Offline []FriendListEntry
	Pending []FriendListEntry
}

func ProjectFriendsList(records []FriendListRecord) FriendsProjection {
	canonical := canonicalizeRecords(records)
	projection := FriendsProjection{}
	for _, record := range canonical {
		entry, include := projectEntry(record)
		if !include {
			continue
		}
		switch entry.Tab {
		case ListTabOnline:
			projection.Online = append(projection.Online, entry)
		case ListTabOffline:
			projection.Offline = append(projection.Offline, entry)
		default:
			projection.Pending = append(projection.Pending, entry)
		}
	}

	sortEntries(projection.Online)
	sortEntries(projection.Offline)
	sortEntries(projection.Pending)
	return projection
}

func canonicalizeRecords(records []FriendListRecord) []FriendListRecord {
	index := make(map[string]FriendListRecord, len(records))
	for _, record := range records {
		identityID := strings.TrimSpace(record.IdentityID)
		if identityID == "" {
			continue
		}
		record.IdentityID = identityID
		existing, ok := index[identityID]
		if !ok {
			index[identityID] = record
			continue
		}
		if shouldReplaceRecord(existing, record) {
			index[identityID] = record
		}
	}
	out := make([]FriendListRecord, 0, len(index))
	for _, record := range index {
		out = append(out, record)
	}
	return out
}

func shouldReplaceRecord(existing, candidate FriendListRecord) bool {
	if candidate.LastUpdatedUnix != existing.LastUpdatedUnix {
		return candidate.LastUpdatedUnix > existing.LastUpdatedUnix
	}
	if candidate.RequestState != existing.RequestState {
		return string(candidate.RequestState) > string(existing.RequestState)
	}
	return string(candidate.PresenceState) > string(existing.PresenceState)
}

func projectEntry(record FriendListRecord) (FriendListEntry, bool) {
	identityID := strings.TrimSpace(record.IdentityID)
	if identityID == "" {
		return FriendListEntry{Reason: ProjectionReasonInvalidIdentity}, false
	}
	entry := FriendListEntry{
		IdentityID:    identityID,
		RequestState:  record.RequestState,
		PresenceState: record.PresenceState,
	}
	if record.RequestState == RequestStateAccepted {
		if isOnlinePresence(record.PresenceState) {
			entry.Tab = ListTabOnline
			entry.Reason = ProjectionReasonAcceptedOnline
			return entry, true
		}
		entry.Tab = ListTabOffline
		entry.Reason = ProjectionReasonAcceptedOffline
		return entry, true
	}
	entry.Tab = ListTabPending
	entry.Reason = ProjectionReasonPending
	return entry, true
}

func isOnlinePresence(state presence.State) bool {
	switch state {
	case presence.StateOnline, presence.StateIdle, presence.StateDND:
		return true
	default:
		return false
	}
}

func sortEntries(entries []FriendListEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		left, right := entries[i], entries[j]
		leftRank := requestStateSortRank(left.RequestState)
		rightRank := requestStateSortRank(right.RequestState)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		leftPresence := presenceSortRank(left.PresenceState)
		rightPresence := presenceSortRank(right.PresenceState)
		if leftPresence != rightPresence {
			return leftPresence < rightPresence
		}
		return strings.ToLower(left.IdentityID) < strings.ToLower(right.IdentityID)
	})
}

func requestStateSortRank(state RequestState) int {
	switch state {
	case RequestStatePending:
		return 0
	case RequestStateAccepted:
		return 1
	case RequestStateBlocked:
		return 2
	case RequestStateDeclined:
		return 3
	case RequestStateCanceled:
		return 4
	default:
		return 5
	}
}

func presenceSortRank(state presence.State) int {
	switch state {
	case presence.StateOnline:
		return 0
	case presence.StateIdle:
		return 1
	case presence.StateDND:
		return 2
	case presence.StateInvisible:
		return 3
	case presence.StateOffline:
		return 4
	default:
		return 5
	}
}

type UIState string

const (
	UIStateLoading       UIState = "loading"
	UIStateEmpty         UIState = "empty"
	UIStateReady         UIState = "ready"
	UIStateError         UIState = "error"
	UIStatePendingAction UIState = "pending-action"
)

type UIEvent string

const (
	UIEventSyncStarted          UIEvent = "sync-started"
	UIEventSyncCompleted        UIEvent = "sync-completed"
	UIEventSyncFailed           UIEvent = "sync-failed"
	UIEventPendingActionQueued  UIEvent = "pending-action-queued"
	UIEventPendingActionCleared UIEvent = "pending-action-cleared"
	UIEventDataCleared          UIEvent = "data-cleared"
)

type UIAction string

const (
	UIActionWait          UIAction = "wait"
	UIActionRetry         UIAction = "retry"
	UIActionAddFriends    UIAction = "add-friends"
	UIActionReviewPending UIAction = "review-pending"
	UIActionNone          UIAction = "none"
)

type UISummary struct {
	OnlineCount  int
	OfflineCount int
	PendingCount int
}

type UIStateDecision struct {
	State  UIState
	Action UIAction
	Reason string
}

func TransitionUIState(current UIState, event UIEvent, summary UISummary) UIStateDecision {
	_ = current
	hasEntries := summary.OnlineCount+summary.OfflineCount > 0
	hasPending := summary.PendingCount > 0

	switch event {
	case UIEventSyncStarted:
		return UIStateDecision{State: UIStateLoading, Action: UIActionWait, Reason: "sync-started"}
	case UIEventSyncFailed:
		return UIStateDecision{State: UIStateError, Action: UIActionRetry, Reason: "sync-failed"}
	case UIEventPendingActionQueued:
		return UIStateDecision{State: UIStatePendingAction, Action: UIActionReviewPending, Reason: "pending-action-queued"}
	case UIEventPendingActionCleared, UIEventSyncCompleted:
		if hasPending {
			return UIStateDecision{State: UIStatePendingAction, Action: UIActionReviewPending, Reason: "pending-action-present"}
		}
		if hasEntries {
			return UIStateDecision{State: UIStateReady, Action: UIActionNone, Reason: "has-friends"}
		}
		return UIStateDecision{State: UIStateEmpty, Action: UIActionAddFriends, Reason: "empty-list"}
	case UIEventDataCleared:
		return UIStateDecision{State: UIStateEmpty, Action: UIActionAddFriends, Reason: "data-cleared"}
	default:
		return UIStateDecision{State: UIStateEmpty, Action: UIActionAddFriends, Reason: "default"}
	}
}
