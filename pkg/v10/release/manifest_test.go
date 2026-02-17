package release

import "testing"

func TestManifestChecklistIncludesSections(t *testing.T) {
	t.Parallel()

	checklist := ManifestChecklist()
	if checklist["landing"] == "" {
		t.Fatal("expected landing entry in manifest checklist")
	}
	compliance := DistributionCompliance()
	if _, ok := compliance["Google Play"]; !ok {
		t.Fatal("expected compliance entry for Google Play")
	}
}
