package durability

import "fmt"

// DurabilityState represents the high-level durability health of a replica set.
type DurabilityState string

const (
	// DurabilityStateHealthy indicates the replica target is satisfied.
	DurabilityStateHealthy DurabilityState = "DURABILITY_HEALTHY"
	// DurabilityStateDegraded indicates the replica target is temporarily unmet but service can continue.
	DurabilityStateDegraded DurabilityState = "DURABILITY_DEGRADED"
	// DurabilityStateUnavailable indicates the replica target is not met and service is at risk.
	DurabilityStateUnavailable DurabilityState = "DURABILITY_UNAVAILABLE"
)

// DurabilityReason provides deterministic reason codes that explain why a durability state was chosen.
type DurabilityReason string

const (
	// ReasonTargetMet is returned when the configured replica target is satisfied.
	ReasonTargetMet DurabilityReason = "REPLICA_TARGET_MET"
	// ReasonTargetUnmet is returned when the replica target is not reached but the service still meets the minimum acceptance threshold.
	ReasonTargetUnmet DurabilityReason = "REPLICA_TARGET_UNMET"
	// ReasonPartialFailure is returned when the minimum acceptance threshold is below the configured lower bound.
	ReasonPartialFailure DurabilityReason = "REPLICA_PARTIAL_FAILURE"
	// ReasonChurnDetected is returned when rolling updates or partial failures are still resolving even though the numeric target is met.
	ReasonChurnDetected DurabilityReason = "REPLICA_CHURN_DETECTED"
)

// DurabilityStatus captures the evaluated durability health, reason, and payload metadata for UI consumers.
type DurabilityStatus struct {
	State          DurabilityState  `json:"state"`
	Reason         DurabilityReason `json:"reason"`
	TargetReplicas int              `json:"targetReplicas"`
	ReadyReplicas  int              `json:"readyReplicas"`
	MinReplicas    int              `json:"minReplicas"`
	ChurnDetected  bool             `json:"churnDetected"`
	Label          string           `json:"label"`
}

// UILabel returns the human-friendly label intended for UI durability badges.
func (ds DurabilityStatus) UILabel() string {
	if ds.Label != "" {
		return ds.Label
	}
	return fmt.Sprintf("Durability %s", ds.State)
}

type replicaState struct {
	online      bool
	transitions int
	initialized bool
}

// ReplicaEvent captures a health change for a replica node.
type ReplicaEvent struct {
	NodeID string
	Online bool
}

// ReplicaAccounting tracks replica states and churn to evaluate durability deterministically.
type ReplicaAccounting struct {
	target           int
	minAcceptable    int
	churnThreshold   int
	nodes            map[string]*replicaState
	transitionWindow int
}

// NewReplicaAccounting creates a ReplicaAccounting instance.
// minAcceptable must be between 0 and target inclusive.
func NewReplicaAccounting(target, minAcceptable, churnThreshold int) *ReplicaAccounting {
	if target < 0 {
		target = 0
	}
	if minAcceptable < 0 {
		minAcceptable = 0
	}
	if minAcceptable > target {
		minAcceptable = target
	}
	if churnThreshold < 1 {
		churnThreshold = 1
	}
	return &ReplicaAccounting{
		target:           target,
		minAcceptable:    minAcceptable,
		churnThreshold:   churnThreshold,
		nodes:            make(map[string]*replicaState),
		transitionWindow: 0,
	}
}

// ApplyEvent records a node health change and increments the churn window if necessary.
func (ra *ReplicaAccounting) ApplyEvent(evt ReplicaEvent) {
	node, ok := ra.nodes[evt.NodeID]
	if !ok {
		node = &replicaState{}
		ra.nodes[evt.NodeID] = node
	}
	if !node.initialized {
		node.online = evt.Online
		node.initialized = true
		return
	}
	if node.online != evt.Online {
		node.online = evt.Online
		node.transitions++
		ra.transitionWindow++
	}
}

// Snapshot produces the current DurabilityStatus without mutating the accounting state.
func (ra *ReplicaAccounting) Snapshot() DurabilityStatus {
	readyCount := 0
	transitionTotal := 0
	for _, node := range ra.nodes {
		if node.online {
			readyCount++
		}
		transitionTotal += node.transitions
	}
	churnActive := transitionTotal >= ra.churnThreshold
	if ra.transitionWindow > 0 && transitionTotal >= ra.churnThreshold {
		ra.transitionWindow = 0
	}
	status := DurabilityStatus{
		TargetReplicas: ra.target,
		ReadyReplicas:  readyCount,
		MinReplicas:    ra.minAcceptable,
		ChurnDetected:  churnActive,
	}
	if readyCount >= ra.target {
		status.State = DurabilityStateHealthy
		status.Reason = ReasonTargetMet
		status.Label = fmt.Sprintf("Durable (%d/%d replicas ready)", readyCount, ra.target)
		if churnActive {
			status.State = DurabilityStateDegraded
			status.Reason = ReasonChurnDetected
			status.Label = fmt.Sprintf("Durability degraded: churn affecting %d nodes", transitionTotal)
		}
		return status
	}
	if readyCount >= ra.minAcceptable {
		status.State = DurabilityStateDegraded
		status.Reason = ReasonTargetUnmet
		status.Label = fmt.Sprintf("Durability degraded: %d/%d replicas ready", readyCount, ra.target)
		return status
	}
	status.State = DurabilityStateUnavailable
	status.Reason = ReasonPartialFailure
	status.Label = fmt.Sprintf("Durability unavailable: only %d/%d replicas ready", readyCount, ra.target)
	return status
}
