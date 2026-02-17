package sfu

import "testing"

func TestCascadingTiersBounds(t *testing.T) {
	t.Parallel()

	if tiers := CascadingTiers(0); tiers != nil {
		t.Fatalf("expected nil tiers for zero participants")
	}
	tiers := CascadingTiers(300)
	if len(tiers) != 3 {
		t.Fatalf("tier count = %d, want 3", len(tiers))
	}
	if tiers[0].Capacity > 64 {
		t.Fatalf("edge capacity = %d, want <=64", tiers[0].Capacity)
	}
	if tiers[2].Capacity < 256 {
		t.Fatalf("core capacity = %d, want >=256", tiers[2].Capacity)
	}
}

func TestSelectForwardingPathPicksMaxCapacity(t *testing.T) {
	t.Parallel()

	tiers := []TopologyTier{
		{Name: "Edge", TierLevel: 1, Capacity: 50},
		{Name: "Regional", TierLevel: 2, Capacity: 80},
		{Name: "Core", TierLevel: 3, Capacity: 60},
	}
	decision := SelectForwardingPath(tiers, 160)
	if decision.Tier.Name != "Regional" {
		t.Fatalf("selected tier = %s, want Regional", decision.Tier.Name)
	}
	if decision.Priority != "degraded" {
		t.Fatalf("priority = %s, want degraded", decision.Priority)
	}
}

func TestFailoverDecisionScenarios(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name              string
		upstreamAvailable bool
		overload          bool
		wantAction        string
	}{
		{name: "upstream down", upstreamAvailable: false, overload: false, wantAction: "promote lower tier"},
		{name: "overload", upstreamAvailable: true, overload: true, wantAction: "split/merge segments"},
		{name: "healthy", upstreamAvailable: true, overload: false, wantAction: "remain"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			decision := FailoverDecision(c.upstreamAvailable, c.overload)
			if decision.Action != c.wantAction {
				t.Fatalf("action = %s, want %s", decision.Action, c.wantAction)
			}
		})
	}
}
