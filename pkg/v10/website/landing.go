package website

// LandingClaims returns the deterministic landing content claims + evidence.
func LandingClaims() map[string]string {
	return map[string]string{
		"reproducible_builds": "pkg/v10/repro/verification.go",
		"global_bootstrap":    "pkg/v10/store/store.go",
		"relay_governance":    "pkg/v10/relay/policy.go",
	}
}
