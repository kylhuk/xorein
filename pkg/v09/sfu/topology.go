package sfu

import "fmt"

// TopologyTier describes SFU cascade tiers.
type TopologyTier struct {
	Name      string
	TierLevel int
	Capacity  int
}

// CascadingTiers returns tier definitions for a target participant count.
func CascadingTiers(targetParticipants int) []TopologyTier {
	if targetParticipants <= 0 {
		return nil
	}
	return []TopologyTier{
		{Name: "Edge", TierLevel: 1, Capacity: min(targetParticipants, 64)},
		{Name: "Regional", TierLevel: 2, Capacity: min(targetParticipants/2, 128)},
		{Name: "Core", TierLevel: 3, Capacity: max(256, targetParticipants/4)},
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ForwardingDecision describes the selected media path.
type ForwardingDecision struct {
	Tier     TopologyTier
	Path     string
	Priority string
}

// SelectForwardingPath picks a tier based on latency heuristics.
func SelectForwardingPath(tiers []TopologyTier, latency int) ForwardingDecision {
	if len(tiers) == 0 {
		return ForwardingDecision{Path: "unknown", Priority: "low"}
	}
	candidate := tiers[0]
	for _, tier := range tiers {
		if tier.Capacity > candidate.Capacity {
			candidate = tier
		}
	}
	priority := "normal"
	if latency > 150 {
		priority = "degraded"
	}
	return ForwardingDecision{Tier: candidate, Path: fmt.Sprintf("tier-%d", candidate.TierLevel), Priority: priority}
}

// DegradationDecision captures failover guidance.
type DegradationDecision struct {
	Tier   string
	Action string
	Reason string
}

// FailoverDecision provides deterministic degrade guidance.
func FailoverDecision(upstreamAvailable bool, overload bool) DegradationDecision {
	if !upstreamAvailable {
		return DegradationDecision{Tier: "fallback", Action: "promote lower tier", Reason: "upstream-unreachable"}
	}
	if overload {
		return DegradationDecision{Tier: "current", Action: "split/merge segments", Reason: "overload"}
	}
	return DegradationDecision{Tier: "current", Action: "remain", Reason: "healthy"}
}
