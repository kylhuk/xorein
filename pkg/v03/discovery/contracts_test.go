package discovery

import "testing"

func TestResolveEntryStateTransitions(t *testing.T) {
	tests := []struct {
		name   string
		last   int64
		expect DirectoryEntryState
	}{
		{"withdrawn", 0, EntryStateWithdrawn},
		{"stale", -10, EntryStateStale},
		{"active", 1, EntryStateActive},
	}

	for _, tt := range tests {
		if got := ResolveEntryState(tt.last); got != tt.expect {
			t.Fatalf("ResolveEntryState(%d) = %q, want %q", tt.last, got, tt.expect)
		}
	}
}

func TestDegradedBrowseHint(t *testing.T) {
	if got := DegradedBrowseHint(0); got != "No stale data tolerance configured" {
		t.Fatalf("hint for zero threshold = %q", got)
	}
	if got := DegradedBrowseHint(500); got != "Stale listings older than 500ms considered degraded" {
		t.Fatalf("hint for positive threshold = %q", got)
	}
}

func TestInvitePolicyDisclosure(t *testing.T) {
	tests := []struct {
		reason   InviteLifecycleReason
		expected string
	}{
		{InviteReasonPolicyBlock, "Invite rejected: policy guardrails prevent broadcast"},
		{InviteReasonRequest, "Request-to-join recorded and queued for response"},
		{InviteReasonCreate, "Invite issued; deterministic acknowledgement follows"},
		{InviteLifecycleReason("other"), "Invite issued; deterministic acknowledgement follows"},
	}

	for _, tt := range tests {
		if got := InvitePolicyDisclosure(tt.reason); got != tt.expected {
			t.Fatalf("InvitePolicyDisclosure(%q) = %q", tt.reason, got)
		}
	}
}
