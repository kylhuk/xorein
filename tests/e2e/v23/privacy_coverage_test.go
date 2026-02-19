package v23

import (
	"testing"

	"github.com/aether/code_aether/pkg/v23/security"
)

func TestPrivacyCoverageBlocksKeywordBackfill(t *testing.T) {
	cfg := security.PrivacyConfig{}
	req := security.BackfillRequest{SpaceID: "space", ChannelID: "chan", Query: "keyword"}
	if err := cfg.ValidateBackfillRequest(req); err == nil {
		t.Fatalf("expected keyword request to be aborted by default")
	}
}

func TestCoverageLabelShowsPartialWhenGaps(t *testing.T) {
	state := security.CoverageState{Gaps: []security.CoverageGap{{Start: 0, End: 5, Reason: "missing"}}}
	if label := security.LabelCoverage(state); label == security.CoverageLabelFull {
		t.Fatalf("label should not claim full coverage when gaps exist")
	}
}
