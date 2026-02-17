package retention

import "testing"

func TestResolveRetentionPolicy(t *testing.T) {
	policy := ResolveRetentionPolicy([]ServerTier{TierEdge})
	if policy.Tier != TierEdge || policy.Days != 14 {
		t.Fatalf("expected edge tier policy, got %v", policy)
	}
	corePolicy := ResolveRetentionPolicy([]ServerTier{TierBurst, TierCore})
	if corePolicy.Tier != TierCore {
		t.Fatalf("expected core tier priority, got %v", corePolicy.Tier)
	}
}

func TestDetermineTransition(t *testing.T) {
	current := RetentionPolicy{Tier: TierEdge, Days: 14}
	target := RetentionPolicy{Tier: TierBurst, Days: 7}
	if DetermineTransition(current, target) != TransitionPurge {
		t.Fatalf("expected purge transition")
	}
	target = RetentionPolicy{Tier: TierCore, Days: 30}
	if DetermineTransition(current, target) != TransitionArchive {
		t.Fatalf("expected archive transition")
	}
}

func TestBuildPurgePlan(t *testing.T) {
	plan := BuildPurgePlan(RetentionPolicy{Days: 5})
	if plan.Reason != "burst purge" {
		t.Fatalf("expected burst purge reason")
	}
}
