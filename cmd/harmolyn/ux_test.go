package main

import (
	"strings"
	"testing"
)

func TestCoverageLabelIncludesGaps(t *testing.T) {
	state := CoverageState{
		LocalStart: "00:00",
		LocalEnd:   "01:00",
		Gaps: []CoverageGap{
			{Start: "00:30", End: "00:40", Reason: ReasonNeedsBackfill},
		},
	}

	label := CoverageLabel(state)
	if !strings.Contains(label, "00:30-00:40") || !strings.Contains(label, string(ReasonNeedsBackfill)) {
		t.Fatalf("unexpected coverage label: %s", label)
	}
}

func TestMissingHistoryReasonLabel(t *testing.T) {
	label := MissingHistoryReasonLabel(ReasonAdjustRange)
	if label != "Adjust the requested range to available coverage" {
		t.Fatalf("unexpected reason label: %s", label)
	}
}

func TestRecommendedBackfillNextAction(t *testing.T) {
	adjustState := CoverageState{Gaps: []CoverageGap{{Reason: ReasonAdjustRange}}}
	if RecommendedBackfillNextAction(adjustState) != ActionAdjustRange {
		t.Fatalf("expected adjust range action")
	}

	fillState := CoverageState{Gaps: []CoverageGap{{Reason: ReasonNeedsBackfill}}}
	if RecommendedBackfillNextAction(fillState) != ActionStartBackfill {
		t.Fatalf("expected start backfill action")
	}

	emptyState := CoverageState{}
	if RecommendedBackfillNextAction(emptyState) != ActionCancelBackfill {
		t.Fatalf("expected cancel action")
	}
}

func TestBackfillActionLabel(t *testing.T) {
	if BackfillActionLabel(ActionStartBackfill) != "Start backfill" {
		t.Fatalf("unexpected label")
	}
}
