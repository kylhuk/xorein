package discovery

import "fmt"

type DirectoryEntryState string

const (
	EntryStateActive    DirectoryEntryState = "active"
	EntryStateStale     DirectoryEntryState = "stale"
	EntryStateWithdrawn DirectoryEntryState = "withdrawn"
)

type InviteLifecycleReason string

const (
	InviteReasonCreate      InviteLifecycleReason = "create"
	InviteReasonRequest     InviteLifecycleReason = "request"
	InviteReasonPolicyBlock InviteLifecycleReason = "policy-block"
)

func ResolveEntryState(lastUpdateMs int64) DirectoryEntryState {
	if lastUpdateMs == 0 {
		return EntryStateWithdrawn
	}
	if lastUpdateMs < 0 {
		return EntryStateStale
	}
	return EntryStateActive
}

func DegradedBrowseHint(staleThresholdMs int64) string {
	if staleThresholdMs <= 0 {
		return "No stale data tolerance configured"
	}
	return fmt.Sprintf("Stale listings older than %dms considered degraded", staleThresholdMs)
}

func InvitePolicyDisclosure(reason InviteLifecycleReason) string {
	switch reason {
	case InviteReasonPolicyBlock:
		return "Invite rejected: policy guardrails prevent broadcast"
	case InviteReasonRequest:
		return "Request-to-join recorded and queued for response"
	default:
		return "Invite issued; deterministic acknowledgement follows"
	}
}
