package v09perf

import (
	"sort"
	"testing"

	"github.com/aether/code_aether/pkg/v09/mobile"
	"github.com/aether/code_aether/pkg/v09/relay"
	"github.com/aether/code_aether/pkg/v09/scale"
	"github.com/aether/code_aether/pkg/v09/sfu"
)

func TestIncrementalScaleCampaign(t *testing.T) {
	t.Parallel()

	const (
		startMembers = 1000
		endMembers   = 1350
		step         = 50
	)

	prevFanouts := []int{0, 0, 0}
	mode := scale.ModeStandard
	prevShardCount := 0
	sawHardened := false
	sawHysteresis := false

	for members := startMembers; members <= endMembers; members += step {
		plan := scale.HierarchyPlan(members)
		if len(plan) != 3 {
			t.Fatalf("hierarchy plan len = %d, want 3", len(plan))
		}

		expectedFanouts := []int{
			clampInt(members, 64),
			clampInt(members/2, 32),
			clampInt(members/4, 16),
		}
		for i, lvl := range plan {
			if lvl.Fanout != expectedFanouts[i] {
				t.Fatalf("tier[%d] fanout = %d, want %d", i, lvl.Fanout, expectedFanouts[i])
			}
			if lvl.Fanout < prevFanouts[i] {
				t.Fatalf("tier[%d] fanout decreased from %d to %d at members=%d", i, prevFanouts[i], lvl.Fanout, members)
			}
			prevFanouts[i] = lvl.Fanout
		}

		mode = scale.NextSecurityMode(mode, members)
		if mode == scale.ModeHardened {
			sawHardened = true
		}
		if sawHardened && mode == scale.ModeStandard && members >= 1200 && members < 1300 {
			sawHysteresis = true
		}

		expectedShardCount := 1 + members/500
		if mode == scale.ModeHardened {
			expectedShardCount++
		}
		shard := scale.ShardingPlan(members, mode)
		if shard.ShardCount != expectedShardCount {
			t.Fatalf("shard count = %d, want %d (mode=%s,members=%d)", shard.ShardCount, expectedShardCount, mode, members)
		}
		expectedMembersPerShard := members / shard.ShardCount
		if expectedMembersPerShard < 1 {
			expectedMembersPerShard = 1
		}
		if shard.MembersPerShard != expectedMembersPerShard {
			t.Fatalf("members per shard = %d, want %d", shard.MembersPerShard, expectedMembersPerShard)
		}
		if shard.ShardCount < prevShardCount && !(mode == scale.ModeStandard && members < 1300) {
			t.Fatalf("shard count dropped unexpectedly from %d to %d at members=%d", prevShardCount, shard.ShardCount, members)
		}
		prevShardCount = shard.ShardCount
	}

	if !sawHardened {
		t.Fatalf("never entered hardened mode")
	}
	if !sawHysteresis {
		t.Fatalf("hysteresis regression not observed")
	}
}

func TestDeterministicScaleFanout(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		members int
		want    []int
	}{
		{name: "large cluster", members: 1200, want: []int{64, 32, 16}},
		{name: "medium cluster", members: 100, want: []int{64, 32, 16}},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			plan := scale.HierarchyPlan(c.members)
			if len(plan) != 3 {
				t.Fatalf("plan len = %d, want 3", len(plan))
			}
			for i, wantFanout := range c.want {
				if plan[i].Fanout != wantFanout {
					t.Fatalf("tier[%d] fanout = %d, want %d", i, plan[i].Fanout, wantFanout)
				}
			}
		})
	}
}

func TestDeterministicRelayOverloadPriority(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		util int
		want []relay.OverloadPriority
	}{
		{name: "stress lane", util: 95, want: []relay.OverloadPriority{relay.PriorityControl, relay.PriorityActive, relay.PriorityStore, relay.PriorityBulk}},
		{name: "moderate lane", util: 80, want: []relay.OverloadPriority{relay.PriorityControl, relay.PriorityActive, relay.PriorityStore}},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			policy := relay.OverloadPolicy(c.util)
			if len(policy) != len(c.want) {
				t.Fatalf("policy len = %d, want %d", len(policy), len(c.want))
			}
			for i, lane := range policy {
				if lane != c.want[i] {
					t.Fatalf("priority[%d] = %s, want %s", i, lane, c.want[i])
				}
			}
		})
	}
}

func TestDeterministicVoiceCascade(t *testing.T) {
	t.Parallel()

	tiers := sfu.CascadingTiers(50)
	if len(tiers) != 3 {
		t.Fatalf("cascade tier count = %d, want 3", len(tiers))
	}
	expectedCaps := []int{50, 25, 256}
	for i, tier := range tiers {
		if tier.Capacity != expectedCaps[i] {
			t.Fatalf("tier[%d] capacity = %d, want %d", i, tier.Capacity, expectedCaps[i])
		}
	}

	decision := sfu.SelectForwardingPath(tiers, 140)
	if decision.Tier.Name != "Core" {
		t.Fatalf("decision tier = %q, want Core", decision.Tier.Name)
	}
	if decision.Path != "tier-3" {
		t.Fatalf("decision path = %s, want tier-3", decision.Path)
	}
	if decision.Priority != "normal" {
		t.Fatalf("priority = %s, want normal", decision.Priority)
	}

	degraded := sfu.SelectForwardingPath(tiers, 180)
	if degraded.Priority != "degraded" {
		t.Fatalf("expected degraded priority for 180ms latency, got %s", degraded.Priority)
	}
}

func TestLatencyThresholdHelper(t *testing.T) {
	t.Parallel()

	t.Run("pass", func(t *testing.T) {
		t.Parallel()
		if medianLatency([]int{180, 195, 205, 190, 200}) > 200 {
			t.Fatalf("median latency unexpectedly above threshold")
		}
	})

	t.Run("regression", func(t *testing.T) {
		t.Parallel()
		if medianLatency([]int{210, 220, 230}) <= 200 {
			t.Fatalf("regression detected: median latency should exceed 200")
		}
	})
}

func TestDeterministicMobileBatteryPolicy(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                      string
		cpu                       int
		battery                   int
		pendingHighPriority       bool
		networkActive             bool
		wantBudget                mobile.BudgetClass
		wantWakeSuppressionReason string
		wantOptimization          string
	}{
		{name: "hard limit", cpu: 90, battery: 15, pendingHighPriority: false, networkActive: true, wantBudget: mobile.BudgetCritical, wantWakeSuppressionReason: "low battery", wantOptimization: "reduce render quality"},
		{name: "high priority allowance", cpu: 50, battery: 50, pendingHighPriority: true, networkActive: true, wantBudget: mobile.BudgetInteractive, wantWakeSuppressionReason: "high-priority", wantOptimization: "stay active"},
		{name: "background idle", cpu: 20, battery: 80, pendingHighPriority: false, networkActive: false, wantBudget: mobile.BudgetBackground, wantWakeSuppressionReason: "budget available", wantOptimization: "suppress syncs"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := mobile.BackgroundBudget(c.cpu, c.battery); got != c.wantBudget {
				t.Fatalf("budget = %s, want %s", got, c.wantBudget)
			}
			wake := mobile.EvaluateWakePolicy(c.battery, c.pendingHighPriority)
			if wake.SuppressionReason != c.wantWakeSuppressionReason {
				t.Fatalf("wake suppression reason = %s, want %s", wake.SuppressionReason, c.wantWakeSuppressionReason)
			}
			if got := mobile.BatteryOptimizationDecision(c.cpu, c.networkActive); got != c.wantOptimization {
				t.Fatalf("optimization decision = %s, want %s", got, c.wantOptimization)
			}
		})
	}
}

func clampInt(value, limit int) int {
	if value < limit {
		return value
	}
	return limit
}

func medianLatency(samples []int) int {
	if len(samples) == 0 {
		return 0
	}
	sorted := append([]int(nil), samples...)
	sort.Ints(sorted)
	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}
