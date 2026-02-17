package scale

import "testing"

func TestHierarchyPlanBounds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		total        int
		wantFanout   int
		wantLen      int
		wantEdgeName string
	}{
		{name: "zero members", total: 0, wantLen: 0},
		{name: "limited members", total: 30, wantLen: 3, wantFanout: 30, wantEdgeName: "Edge"},
		{name: "large members", total: 5_000, wantLen: 3, wantFanout: 64, wantEdgeName: "Edge"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			plan := HierarchyPlan(c.total)
			if len(plan) != c.wantLen {
				t.Fatalf("plan len = %d, want %d", len(plan), c.wantLen)
			}
			if c.wantLen == 0 {
				return
			}
			if plan[0].Name != c.wantEdgeName {
				t.Fatalf("first tier name = %s, want %s", plan[0].Name, c.wantEdgeName)
			}
			if plan[0].Fanout != c.wantFanout {
				t.Fatalf("edge fanout = %d, want %d", plan[0].Fanout, c.wantFanout)
			}
		})
	}
}

func TestLazyLoadPlanPercentiles(t *testing.T) {
	t.Parallel()

	plan := LazyLoadPlan(450)
	if len(plan) != 3 {
		t.Fatalf("plan len = %d, want 3", len(plan))
	}
	if plan[2].TriggerPercentile > 100 {
		t.Fatalf("passive percentile = %d, cannot exceed 100", plan[2].TriggerPercentile)
	}
}

func TestNextSecurityModeTransitions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		current SecurityMode
		load    int
		want    SecurityMode
	}{
		{name: "escalate to hardened", current: ModeStandard, load: 1250, want: ModeHardened},
		{name: "stay hardened", current: ModeHardened, load: 1300, want: ModeHardened},
		{name: "deescalate to standard", current: ModeHardened, load: 100, want: ModeStandard},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if mode := NextSecurityMode(c.current, c.load); mode != c.want {
				t.Fatalf("mode = %s, want %s", mode, c.want)
			}
		})
	}
}

func TestShardingPlanIncludesModeEpoch(t *testing.T) {
	t.Parallel()

	guidance := ShardingPlan(750, ModeHardened)
	if guidance.ShardCount < 2 {
		t.Fatalf("shard count = %d, want >=2", guidance.ShardCount)
	}
	if guidance.MembersPerShard < 1 {
		t.Fatalf("members per shard = %d, want >=1", guidance.MembersPerShard)
	}
	if guidance.ModeEpoch == "" {
		t.Fatalf("mode epoch must be populated")
	}
}
