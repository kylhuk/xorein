package docs

// UserGuideChecklist returns the user guide acceptance entries.
func UserGuideChecklist() []string {
	return []string{"onboarding", "security-modes", "limitations"}
}

// AdminGuideChecklist returns the admin guide chapters.
func AdminGuideChecklist() []string {
	return []string{"moderation", "relay-ops", "alerting"}
}

// DeveloperGuideChecklist returns the developer/ API guide anchors.
func DeveloperGuideChecklist() []string {
	return []string{"architecture", "extensions", "compatibility", "api-taxonomy"}
}
