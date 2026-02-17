package retention

import (
	"testing"
	"time"
)

func TestResolveRetentionPolicyPriority(t *testing.T) {
	policy := ResolveRetentionPolicy([]ServerTier{TierBurst, TierEdge})
	if policy.Tier != TierEdge {
		t.Fatalf("expected edge tier, got %s", policy.Tier)
	}
	policy = ResolveRetentionPolicy([]ServerTier{TierBurst, TierCore, TierEdge})
	if policy.Tier != TierCore {
		t.Fatalf("expected core tier, got %s", policy.Tier)
	}
}

func TestDetermineTransitionAndPurgePlan(t *testing.T) {
	current := RetentionPolicy{Tier: TierEdge, Days: 14}
	target := RetentionPolicy{Tier: TierBurst, Days: 7}
	if action := DetermineTransition(current, target); action != TransitionPurge {
		t.Fatalf("expected purge when tightening, got %s", action)
	}
	plan := BuildPurgePlan(target)
	if plan.Reason != "burst purge" {
		t.Fatalf("unexpected purge reason %s", plan.Reason)
	}
}

func TestPolicyStoreLifecycle(t *testing.T) {
	initial := RetentionPolicy{Tier: TierEdge, Days: 14}
	store := NewPolicyStore(initial, "auditor")
	time.Sleep(1 * time.Nanosecond)
	change := store.ApplyChange(RetentionPolicy{Tier: TierCore, Days: 30}, "auditor")
	audit := store.BuildAuditTrail()
	if len(audit) != 2 {
		t.Fatalf("expected two records, got %d", len(audit))
	}
	if audit[1].ChangeID != change.ChangeID {
		t.Fatalf("current record mismatch")
	}
	if store.AcceptRecord(time.Now().Add(-45 * 24 * time.Hour)) {
		t.Fatalf("expected old record to be rejected")
	}
	if !store.AcceptRecord(time.Now().Add(-time.Hour)) {
		t.Fatalf("expected recent record to be accepted")
	}
	frequency, topic := store.EnforcePurge()
	if topic != "retention.core" {
		t.Fatalf("unexpected topic %s", topic)
	}
	if frequency != 30*24*time.Hour {
		t.Fatalf("unexpected frequency %v", frequency)
	}
	changeEvent := store.DescribeChange("operator")
	if changeEvent.Previous.Tier != TierEdge {
		t.Fatalf("expected previous tier edge, got %s", changeEvent.Previous.Tier)
	}
}
