package security

import "testing"

func TestSecurityArtifactsPresent(t *testing.T) {
	t.Parallel()

	if len(AssetInventory()) < 3 {
		t.Fatal("expected asset inventory to list at least three categories")
	}
	if len(ThreatModel()) < 3 {
		t.Fatal("expected threat model to cover confidentiality/integrity/availability")
	}
	if len(EngagementCriteria()) < 3 {
		t.Fatal("expected engagement criteria to contain independence/expertise/conflict")
	}
	if len(FindingLifecycle()) != 3 {
		t.Fatalf("finding lifecycle length = %d, want 3", len(FindingLifecycle()))
	}
}
