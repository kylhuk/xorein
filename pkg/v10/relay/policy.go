package relay

// ReliabilityScores returns deterministic reliability scores used by the gate.
func ReliabilityScores() map[string]int {
	return map[string]int{
		"bootstrap-node-a": 95,
		"bootstrap-node-b": 92,
		"relay-node-1":     89,
	}
}

// AbuseResponseClass returns deterministic severity classes for abuses.
func AbuseResponseClass() map[string]string {
	return map[string]string{
		"high":   "immediate review",
		"medium": "operator notification",
		"low":    "monitoring",
	}
}
