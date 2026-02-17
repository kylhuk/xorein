package security

// EngagementCriteria returns deterministic auditor engagement entries.
func EngagementCriteria() map[string]string {
	return map[string]string{
		"independence": "no prior contracts",
		"expertise":    "E2EE + networking",
		"conflict":     "public disclosure",
	}
}
