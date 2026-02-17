package retention

type ServerTier string

const (
	TierCore  ServerTier = "core"
	TierEdge  ServerTier = "edge"
	TierBurst ServerTier = "burst"
)

type RetentionPolicy struct {
	Tier ServerTier
	Days int
}

func ResolveRetentionPolicy(tiers []ServerTier) RetentionPolicy {
	choices := map[ServerTier]int{
		TierCore:  30,
		TierEdge:  14,
		TierBurst: 7,
	}
	priority := TierBurst
	for _, tier := range tiers {
		if tier == TierCore {
			priority = TierCore
			break
		}
		if tier == TierEdge && priority != TierCore {
			priority = TierEdge
		}
	}
	return RetentionPolicy{Tier: priority, Days: choices[priority]}
}

type TransitionAction string

const (
	TransitionKeep    TransitionAction = "retain"
	TransitionArchive TransitionAction = "archive"
	TransitionPurge   TransitionAction = "purge"
)

func DetermineTransition(current, target RetentionPolicy) TransitionAction {
	if current.Tier == target.Tier {
		return TransitionKeep
	}
	if target.Days < current.Days {
		return TransitionPurge
	}
	return TransitionArchive
}

type PurgePlan struct {
	DaysUntilPurge int
	Reason         string
}

func BuildPurgePlan(policy RetentionPolicy) PurgePlan {
	if policy.Days <= 7 {
		return PurgePlan{DaysUntilPurge: policy.Days, Reason: "burst purge"}
	}
	return PurgePlan{DaysUntilPurge: policy.Days, Reason: "standard retention"}
}
