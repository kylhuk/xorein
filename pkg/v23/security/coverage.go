package security

// CoverageLabel enumerates the summary state tag for coverage.
type CoverageLabel string

const (
	// CoverageLabelFull is returned only when every segment is present.
	CoverageLabelFull CoverageLabel = "coverage.full_history"
	// CoverageLabelPartial expresses that known gaps exist.
	CoverageLabelPartial CoverageLabel = "coverage.partial_history"
	// CoverageLabelIncomplete means the coverage state is unknown or still gathering.
	CoverageLabelIncomplete CoverageLabel = "coverage.incomplete_history"
)

// CoverageGap describes a contiguous window that is missing.
type CoverageGap struct {
	Start  int64
	End    int64
	Reason string
}

// CoverageState tracks the recorded coverage windows and any missing segments.
type CoverageState struct {
	Gaps     []CoverageGap
	Complete bool
}

// LabelCoverage deterministically maps the state to a coverage label.
func LabelCoverage(state CoverageState) CoverageLabel {
	switch {
	case state.Complete && len(state.Gaps) == 0:
		return CoverageLabelFull
	case len(state.Gaps) > 0:
		return CoverageLabelPartial
	default:
		return CoverageLabelIncomplete
	}
}
