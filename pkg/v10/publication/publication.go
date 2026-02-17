package publication

// SectionMap returns a deterministic map of normative/informative sections for v1.0.
func SectionMap() map[string]string {
	return map[string]string{
		"normative":   "protocol-surface, compatibility, runtime modes",
		"informative": "governance, terminology, transition",
	}
}

// SectionList returns the formatted section keys for validation.
func SectionList() []string {
	return []string{"normative", "informative", "appendices"}
}
