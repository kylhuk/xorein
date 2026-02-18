package ui

import (
	"fmt"

	"github.com/aether/code_aether/pkg/v15/capture"
	"github.com/aether/code_aether/pkg/v15/screenshare"
)

// ControlState exposes deterministic UI control permutations.
type ControlState struct {
	CanStart         bool
	CanStop          bool
	Source           capture.SourceType
	QualityIndicator string
	Hint             string
	AdaptationLayer  int
}

// ComposeControlState builds a deterministic UI state for the given transport status.
func ComposeControlState(source capture.SourceType, state screenshare.TransportState, bandwidthKbps int) ControlState {
	decision := screenshare.DetermineAdaptation(bandwidthKbps)
	hint := screenshare.RecoveryHint(state)
	quality := fmt.Sprintf("q=%dkbps", decision.BitrateKbps)
	switch state {
	case screenshare.StateIdle:
		return ControlState{CanStart: true, Source: source, QualityIndicator: quality, Hint: "Ready to share.", AdaptationLayer: decision.Layer}
	case screenshare.StateConnecting, screenshare.StateActive:
		return ControlState{CanStart: false, CanStop: true, Source: source, QualityIndicator: quality, Hint: hint, AdaptationLayer: decision.Layer}
	case screenshare.StateFallback, screenshare.StateError:
		return ControlState{CanStart: false, CanStop: true, Source: source, QualityIndicator: quality, Hint: hint + " Check recovery.", AdaptationLayer: decision.Layer}
	case screenshare.StatePaused:
		return ControlState{CanStart: true, CanStop: true, Source: source, QualityIndicator: quality, Hint: "Paused: tap resume.", AdaptationLayer: decision.Layer}
	default:
		return ControlState{CanStart: false, CanStop: false, Source: source, QualityIndicator: quality, Hint: "Status unknown.", AdaptationLayer: decision.Layer}
	}
}
