package security

// ThreatModel returns deterministic threat model entries.
func ThreatModel() map[string]string {
	return map[string]string{
		"confidentiality": "mode_epoch_id leakage",
		"integrity":       "downgrade attempts",
		"availability":    "relay saturation",
	}
}
