package relay

// CapacityEnvelope tracks relay capacity distribution.
type CapacityEnvelope struct {
	Limit        int
	Control      int
	ActiveMedia  int
	StoreForward int
	BulkSync     int
}

// UtilizationPercent returns simple utilization as an integer.
func (c CapacityEnvelope) UtilizationPercent() int {
	if c.Limit == 0 {
		if c.Control+c.ActiveMedia+c.StoreForward+c.BulkSync == 0 {
			return 0
		}
		return 100
	}
	used := c.Control + c.ActiveMedia + c.StoreForward + c.BulkSync
	return (used * 100) / c.Limit
}

// OverloadPriority enumerates priority lanes.
type OverloadPriority string

const (
	PriorityControl OverloadPriority = "control"
	PriorityActive  OverloadPriority = "active-media"
	PriorityStore   OverloadPriority = "store-forward"
	PriorityBulk    OverloadPriority = "bulk-sync"
)

// OverloadPolicy returns the priority order used under overload.
func OverloadPolicy(utilPercent int) []OverloadPriority {
	if utilPercent >= 90 {
		return []OverloadPriority{PriorityControl, PriorityActive, PriorityStore, PriorityBulk}
	}
	return []OverloadPriority{PriorityControl, PriorityActive, PriorityStore}
}

// RecoveryPosture describes next steps after overload.
type RecoveryPosture struct {
	NextStep            string
	RecoveryDelayMillis int
}

// RecoveryPlan suggests deterministic recovery behavior.
func RecoveryPlan(utilPercent int) RecoveryPosture {
	if utilPercent >= 95 {
		return RecoveryPosture{NextStep: "throttle bulk sync and escalate", RecoveryDelayMillis: 1500}
	}
	if utilPercent >= 75 {
		return RecoveryPosture{NextStep: "throttle store-forward", RecoveryDelayMillis: 700}
	}
	return RecoveryPosture{NextStep: "steady", RecoveryDelayMillis: 300}
}
