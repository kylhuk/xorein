package security

// FindingLifecycle returns deterministic severity handling.
func FindingLifecycle() map[string]string {
	return map[string]string{
		"critical": "block release + patch schedule",
		"major":    "owner remediation + regression test",
		"minor":    "monitor + document",
	}
}
