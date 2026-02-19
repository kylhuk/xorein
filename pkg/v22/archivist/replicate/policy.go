package replicate

type EndpointID string

type DurabilityHealth string

type ResultReason string

const (
	HealthDurable  DurabilityHealth = "HISTORY_DURABILITY_HEALTHY"
	HealthDegraded DurabilityHealth = "HISTORY_DURABILITY_DEGRADED"

	ResultReplicaTargetUnmet       ResultReason = "REPLICA_TARGET_UNMET"
	ResultReplicaWritePartial      ResultReason = "REPLICA_WRITE_PARTIAL"
	ResultReplicaHealingInProgress ResultReason = "REPLICA_HEALING_IN_PROGRESS"
)

type Policy struct {
	R    int
	RMin int
}

type ReplicationResult struct {
	SuccessCount int
	Health       DurabilityHealth
	Reason       ResultReason
	TargetMet    bool
}

func (p Policy) evaluate(successes int) (DurabilityHealth, ResultReason, bool) {
	if successes >= p.R {
		return HealthDurable, "", true
	}
	if successes >= p.RMin {
		return HealthDegraded, ResultReplicaWritePartial, false
	}
	return HealthDegraded, ResultReplicaTargetUnmet, false
}

func Replicate(policy Policy, targets []EndpointID, writer func(EndpointID) error) ReplicationResult {
	if policy.R <= 0 {
		policy.R = 1
	}
	if policy.RMin <= 0 {
		policy.RMin = 1
	}

	var successes int
	for _, target := range targets {
		if writer == nil {
			continue
		}
		if err := writer(target); err == nil {
			successes++
		}
	}

	health, reason, met := policy.evaluate(successes)
	return ReplicationResult{SuccessCount: successes, Health: health, Reason: reason, TargetMet: met}
}

type HealingResult struct {
	HealedTokens []EndpointID
	Reason       ResultReason
	SuccessTotal int
}

func Heal(policy Policy, known []EndpointID, candidates []EndpointID, writer func(EndpointID) error) HealingResult {
	if policy.R <= 0 {
		policy.R = 1
	}
	missing := policy.R - len(known)
	if missing <= 0 {
		return HealingResult{Reason: ResultReplicaHealingInProgress, SuccessTotal: len(known)}
	}

	var healed []EndpointID
	present := make(map[EndpointID]struct{})
	for _, id := range known {
		present[id] = struct{}{}
	}

	for _, candidate := range candidates {
		if _, ok := present[candidate]; ok {
			continue
		}
		if writer == nil {
			continue
		}
		if err := writer(candidate); err == nil {
			healed = append(healed, candidate)
			present[candidate] = struct{}{}
			if len(healed) >= missing {
				break
			}
		}
	}

	total := len(present)
	reason := ResultReplicaTargetUnmet
	if len(healed) > 0 {
		reason = ResultReplicaHealingInProgress
	}

	return HealingResult{HealedTokens: healed, Reason: reason, SuccessTotal: total}
}
