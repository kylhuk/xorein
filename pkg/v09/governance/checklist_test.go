package governance

import (
	"strings"
	"testing"
)

func TestAdditiveChecklistItems(t *testing.T) {
	t.Parallel()

	items := AdditiveChecklist()
	if len(items) != 3 {
		t.Fatalf("items len = %d, want 3", len(items))
	}
	ids := map[string]struct{}{}
	for _, item := range items {
		ids[item.ID] = struct{}{}
	}
	for _, id := range []string{"GOV-ADD-01", "GOV-ADD-02", "GOV-ADD-03"} {
		if _, ok := ids[id]; !ok {
			t.Fatalf("missing checklist item %s", id)
		}
	}
}

func TestMajorPathTriggerClassifier(t *testing.T) {
	t.Parallel()

	cases := []struct {
		breaking bool
		context  string
		wantSub  string
	}{
		{breaking: false, context: "", wantSub: "additive path"},
		{breaking: true, context: "multistream", wantSub: "new multistream ID"},
		{breaking: true, context: "downgrade", wantSub: "downgrade negotiation"},
		{breaking: true, context: "other", wantSub: "major-path"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.context, func(t *testing.T) {
			t.Parallel()
			result := MajorPathTriggerClassifier(c.breaking, c.context)
			if !strings.Contains(result, c.wantSub) {
				t.Fatalf("result = %q, want substring %q", result, c.wantSub)
			}
		})
	}
}

func TestLicensingStatus(t *testing.T) {
	t.Parallel()

	if got := LicensingStatus("AGPL", "CC-BY-SA"); got != "AGPL / CC-BY-SA" {
		t.Fatalf("licensing = %q", got)
	}
}
