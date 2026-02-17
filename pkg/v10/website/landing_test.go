package website

import "testing"

func TestLandingClaimsReferenceArtifacts(t *testing.T) {
	t.Parallel()

	claims := LandingClaims()
	if claims["global_bootstrap"] != "pkg/v10/store/store.go" {
		t.Fatalf("unexpected global bootstrap claim: %q", claims["global_bootstrap"])
	}
	if claims["relay_governance"] == "" {
		t.Fatal("expected relay governance claim to exist")
	}
}
