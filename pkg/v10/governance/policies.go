package governance

// NamingGovernance returns canonical names for the v1.0 release.
func NamingGovernance() map[string]string {
	return map[string]string{
		"client":  "Harmolyn",
		"backend": "xorein",
		"legacy":  "Aether",
	}
}

// AdditiveChecklist returns a deterministic list of additive-only constraints.
func AdditiveChecklist() []string {
	return []string{"new-fields only", "reserved numbers maintained", "downgrade-safe defaults"}
}

// MajorPathTriggerClassifier returns a descriptive string summarizing triggers.
func MajorPathTriggerClassifier(needsAEP bool, reference string) string {
	status := "minor-path"
	if needsAEP {
		status = "major-path" + ":" + reference
	}
	return status
}
