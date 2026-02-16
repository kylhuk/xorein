package friends

import (
	"testing"

	"github.com/aether/code_aether/pkg/v02/presence"
)

func TestProjectFriendsListSegmentationAndOrder(t *testing.T) {
	records := []FriendListRecord{
		{IdentityID: "zoe", RequestState: RequestStateAccepted, PresenceState: presence.StateDND, LastUpdatedUnix: 10},
		{IdentityID: "amy", RequestState: RequestStateAccepted, PresenceState: presence.StateOnline, LastUpdatedUnix: 11},
		{IdentityID: "ian", RequestState: RequestStateAccepted, PresenceState: presence.StateIdle, LastUpdatedUnix: 12},
		{IdentityID: "ben", RequestState: RequestStateAccepted, PresenceState: presence.StateOffline, LastUpdatedUnix: 13},
		{IdentityID: "carl", RequestState: RequestStatePending, PresenceState: presence.StateOffline, LastUpdatedUnix: 14},
		{IdentityID: "dina", RequestState: RequestStateBlocked, PresenceState: presence.StateOffline, LastUpdatedUnix: 15},
	}

	projected := ProjectFriendsList(records)

	assertIDs(t, projected.Online, []string{"amy", "ian", "zoe"})
	assertIDs(t, projected.Offline, []string{"ben"})
	assertIDs(t, projected.Pending, []string{"carl", "dina"})
}

func TestProjectFriendsListCanonicalizesDuplicateIdentity(t *testing.T) {
	records := []FriendListRecord{
		{IdentityID: "alex", RequestState: RequestStatePending, PresenceState: presence.StateOffline, LastUpdatedUnix: 5},
		{IdentityID: "alex", RequestState: RequestStateAccepted, PresenceState: presence.StateOnline, LastUpdatedUnix: 9},
	}

	projected := ProjectFriendsList(records)
	if len(projected.Online) != 1 || projected.Online[0].IdentityID != "alex" {
		t.Fatalf("expected alex in online tab: %+v", projected)
	}
	if len(projected.Pending) != 0 {
		t.Fatalf("expected no pending entries: %+v", projected.Pending)
	}
}

func TestTransitionUIState(t *testing.T) {
	tests := []struct {
		name    string
		current UIState
		event   UIEvent
		summary UISummary
		want    UIStateDecision
	}{
		{
			name:    "sync started goes loading",
			current: UIStateEmpty,
			event:   UIEventSyncStarted,
			want:    UIStateDecision{State: UIStateLoading, Action: UIActionWait, Reason: "sync-started"},
		},
		{
			name:    "sync failed goes error",
			current: UIStateLoading,
			event:   UIEventSyncFailed,
			want:    UIStateDecision{State: UIStateError, Action: UIActionRetry, Reason: "sync-failed"},
		},
		{
			name:    "pending action queues pending state",
			current: UIStateReady,
			event:   UIEventPendingActionQueued,
			want:    UIStateDecision{State: UIStatePendingAction, Action: UIActionReviewPending, Reason: "pending-action-queued"},
		},
		{
			name:    "sync complete with entries goes ready",
			current: UIStateLoading,
			event:   UIEventSyncCompleted,
			summary: UISummary{OnlineCount: 1},
			want:    UIStateDecision{State: UIStateReady, Action: UIActionNone, Reason: "has-friends"},
		},
		{
			name:    "sync complete with pending goes pending-action",
			current: UIStateLoading,
			event:   UIEventSyncCompleted,
			summary: UISummary{PendingCount: 2},
			want:    UIStateDecision{State: UIStatePendingAction, Action: UIActionReviewPending, Reason: "pending-action-present"},
		},
		{
			name:    "sync complete with no entries goes empty",
			current: UIStateLoading,
			event:   UIEventSyncCompleted,
			want:    UIStateDecision{State: UIStateEmpty, Action: UIActionAddFriends, Reason: "empty-list"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := TransitionUIState(tc.current, tc.event, tc.summary)
			if got != tc.want {
				t.Fatalf("TransitionUIState()=%+v want %+v", got, tc.want)
			}
		})
	}
}

func assertIDs(t *testing.T, entries []FriendListEntry, want []string) {
	t.Helper()
	if len(entries) != len(want) {
		t.Fatalf("entry count=%d want %d entries=%+v", len(entries), len(want), entries)
	}
	for i := range want {
		if entries[i].IdentityID != want[i] {
			t.Fatalf("entry[%d]=%q want %q", i, entries[i].IdentityID, want[i])
		}
	}
}
