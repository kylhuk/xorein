package governance

import (
	"strings"
	"testing"
)

func TestNamingAndChecklistDeterminism(t *testing.T) {
	t.Parallel()

	names := NamingGovernance()
	if names["client"] != "Harmolyn" {
		t.Fatalf("expected client to be Harmolyn, got %q", names["client"])
	}
	if names["backend"] != "xorein" {
		t.Fatalf("expected backend to be xorein, got %q", names["backend"])
	}

	additive := AdditiveChecklist()
	if len(additive) < 3 {
		t.Fatalf("expected additive checklist to have at least 3 entries, got %d", len(additive))
	}

	gotMajor := MajorPathTriggerClassifier(true, "multistream")
	if gotMajor == "minor-path" || !strings.Contains(gotMajor, "major-path") {
		t.Fatalf("unexpected major path classifier output: %q", gotMajor)
	}
}
