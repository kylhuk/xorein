package main

import (
	"fmt"
	"strings"
)

type CoverageGap struct {
	Start  string
	End    string
	Reason MissingHistoryReason
}

type CoverageState struct {
	LocalStart string
	LocalEnd   string
	Gaps       []CoverageGap
}

func CoverageLabel(state CoverageState) string {
	base := fmt.Sprintf("Searched local footage from %s to %s", state.LocalStart, state.LocalEnd)
	if len(state.Gaps) == 0 {
		return base + " with full coverage"
	}

	gapSummaries := make([]string, len(state.Gaps))
	for i, gap := range state.Gaps {
		gapSummaries[i] = fmt.Sprintf("%s-%s (%s)", gap.Start, gap.End, gap.Reason)
	}
	return base + " with coverage gaps: " + strings.Join(gapSummaries, "; ")
}

type MissingHistoryReason string

const (
	ReasonNoData        MissingHistoryReason = "no_local_history"
	ReasonNeedsBackfill MissingHistoryReason = "missing_backfill"
	ReasonAdjustRange   MissingHistoryReason = "adjust_range"
)

func MissingHistoryReasonLabel(reason MissingHistoryReason) string {
	switch reason {
	case ReasonNoData:
		return "No local history; try backfilling"
	case ReasonNeedsBackfill:
		return "Backfill is needed to cover this range"
	case ReasonAdjustRange:
		return "Adjust the requested range to available coverage"
	default:
		return "History coverage unknown"
	}
}

type BackfillAction string

const (
	ActionStartBackfill  BackfillAction = "start_backfill"
	ActionCancelBackfill BackfillAction = "cancel_backfill"
	ActionAdjustRange    BackfillAction = "adjust_range"
)

func RecommendedBackfillNextAction(state CoverageState) BackfillAction {
	if len(state.Gaps) == 0 {
		return ActionCancelBackfill
	}

	for _, gap := range state.Gaps {
		switch gap.Reason {
		case ReasonAdjustRange:
			return ActionAdjustRange
		default:
			return ActionStartBackfill
		}
	}
	return ActionStartBackfill
}

func BackfillActionLabel(action BackfillAction) string {
	switch action {
	case ActionStartBackfill:
		return "Start backfill"
	case ActionAdjustRange:
		return "Adjust search range"
	case ActionCancelBackfill:
		return "Cancel backfill"
	default:
		return "Review coverage settings"
	}
}
