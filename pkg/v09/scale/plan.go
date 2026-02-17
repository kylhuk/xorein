package scale

import "fmt"

// HierarchyLevel captures GossipSub fanout recommendations.
type HierarchyLevel struct {
	Name      string
	TierLevel int
	Fanout    int
}

// HierarchyPlan builds a deterministic hierarchy plan for totalMembers.
func HierarchyPlan(totalMembers int) []HierarchyLevel {
	if totalMembers <= 0 {
		return nil
	}
	return []HierarchyLevel{
		{Name: "Edge", TierLevel: 1, Fanout: min(totalMembers, 64)},
		{Name: "Aggregation", TierLevel: 2, Fanout: min(totalMembers/2, 32)},
		{Name: "Core", TierLevel: 3, Fanout: min(totalMembers/4, 16)},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MemberClass names the lazy-loading classes.
type MemberClass string

const (
	MemberActive  MemberClass = "active"
	MemberNearby  MemberClass = "nearby"
	MemberPassive MemberClass = "passive"
)

// LazyLoadTrigger describes when classes should refresh.
type LazyLoadTrigger struct {
	Class             MemberClass
	TriggerPercentile int
}

// LazyLoadPlan returns load triggers based on totalMembers.
func LazyLoadPlan(totalMembers int) []LazyLoadTrigger {
	base := totalMembers / 100
	if base < 1 {
		base = 1
	}
	return []LazyLoadTrigger{
		{Class: MemberActive, TriggerPercentile: min(100, base*5)},
		{Class: MemberNearby, TriggerPercentile: min(100, base*15)},
		{Class: MemberPassive, TriggerPercentile: min(100, base*30)},
	}
}

// SecurityMode defines scale-driven encryption posture.
type SecurityMode string

const (
	ModeStandard SecurityMode = "standard"
	ModeHardened SecurityMode = "hardened"
)

// SecurityTransition holds hysteresis thresholds.
type SecurityTransition struct {
	Mode           SecurityMode
	EnterThreshold int
	ExitThreshold  int
}

var securityTransitions = []SecurityTransition{
	{Mode: ModeStandard, EnterThreshold: 0, ExitThreshold: 1300},
	{Mode: ModeHardened, EnterThreshold: 1200, ExitThreshold: 0},
}

// NextSecurityMode picks the next mode with hysteresis.
func NextSecurityMode(current SecurityMode, load int) SecurityMode {
	for _, transition := range securityTransitions {
		if transition.Mode == ModeHardened && load >= transition.EnterThreshold {
			return ModeHardened
		}
		if transition.Mode == ModeStandard && current == ModeHardened && load < transition.ExitThreshold {
			return ModeStandard
		}
	}
	return current
}

// ShardingGuidance describes shard counts.
type ShardingGuidance struct {
	ShardCount      int
	MembersPerShard int
	ModeEpoch       string
	Recommendation  string
}

// ShardingPlan returns deterministic sharding guidance.
func ShardingPlan(members int, mode SecurityMode) ShardingGuidance {
	shardCount := 1 + members/500
	if mode == ModeHardened {
		shardCount += 1
	}
	perShard := max(1, members/shardCount)
	epoch := fmt.Sprintf("mode-%s-%04d", mode, members)
	return ShardingGuidance{
		ShardCount:      shardCount,
		MembersPerShard: perShard,
		ModeEpoch:       epoch,
		Recommendation:  "Shard names incorporate the mode epoch to enforce deterministic mode transitions.",
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
