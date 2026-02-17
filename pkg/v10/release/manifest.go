package release

// ManifestChecklist returns the release bundle mapping.
func ManifestChecklist() map[string]string {
	return map[string]string{"doc-suite": "VA-D1..D6", "landing": "VA-W1..W3", "app-distribution": "VA-A1..A3"}
}

// DistributionCompliance returns per-platform compliance statements.
func DistributionCompliance() map[string]string {
	return map[string]string{
		"Google Play": "security+compatibility",
		"App Store":   "privacy+review",
		"Microsoft":   "quality metadata",
		"Flathub":     "audit-ready",
	}
}
