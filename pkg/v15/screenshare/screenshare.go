package screenshare

import "fmt"

// TransportState captures deterministic transport state progression.
type TransportState string

const (
	StateIdle       TransportState = "idle"
	StateConnecting TransportState = "connecting"
	StateActive     TransportState = "active"
	StatePaused     TransportState = "paused"
	StateFallback   TransportState = "fallback"
	StateError      TransportState = "error"
)

// TransportEvent lists events that drive transitions.
type TransportEvent string

const (
	EventStart       TransportEvent = "start"
	EventConnected   TransportEvent = "connected"
	EventQualityDrop TransportEvent = "quality-drop"
	EventRecover     TransportEvent = "recover"
	EventStop        TransportEvent = "stop"
	EventFail        TransportEvent = "fail"
)

// Transition deterministically updates the TransportState for an event.
func Transition(current TransportState, event TransportEvent) TransportState {
	switch event {
	case EventStart:
		if current == StateIdle || current == StatePaused {
			return StateConnecting
		}
		return current
	case EventConnected:
		if current == StateConnecting {
			return StateActive
		}
		return current
	case EventQualityDrop:
		if current == StateActive {
			return StateFallback
		}
		if current == StateFallback {
			return StateError
		}
		return current
	case EventRecover:
		if current == StateFallback {
			return StateActive
		}
		if current == StateError {
			return StateFallback
		}
		return current
	case EventStop:
		return StateIdle
	case EventFail:
		return StateError
	default:
		return current
	}
}

// AdaptationDecision describes the deterministic fallback tier.
type AdaptationDecision struct {
	Layer       int
	BitrateKbps int
	Reason      string
}

// DetermineAdaptation picks a layer based on bandwidth headroom.
func DetermineAdaptation(bandwidthKbps int) AdaptationDecision {
	switch {
	case bandwidthKbps >= 4000:
		return AdaptationDecision{Layer: 3, BitrateKbps: 4000, Reason: "nominal"}
	case bandwidthKbps >= 2500:
		return AdaptationDecision{Layer: 2, BitrateKbps: 2500, Reason: "traffic"}
	case bandwidthKbps >= 1200:
		return AdaptationDecision{Layer: 1, BitrateKbps: 1200, Reason: "congestion"}
	default:
		return AdaptationDecision{Layer: 0, BitrateKbps: 600, Reason: "critical"}
	}
}

// RecoveryHint returns a no-limbo message for fallback states.
func RecoveryHint(state TransportState) string {
	switch state {
	case StateFallback:
		return "Network constrained: retry or pause other traffic."
	case StateError:
		return "Transport error: restart screen share once network stabilizes."
	case StatePaused:
		return "Paused: resume when ready."
	default:
		return "Streaming stable."
	}
}

// FallbackTransition provides the fallback path for fallback error reasons.
func FallbackTransition(reason string) TransportState {
	if reason == "network" {
		return StateFallback
	}
	return StateError
}

// Descriptor summarizes the transport/adaptation state for deterministic viewers.
type Descriptor struct {
	State    TransportState
	Decision AdaptationDecision
	Recovery string
}

// Summarize builds a Descriptor from state and observed bandwidth.
func Summarize(state TransportState, bandwidthKbps int) Descriptor {
	decision := DetermineAdaptation(bandwidthKbps)
	return Descriptor{
		State:    state,
		Decision: decision,
		Recovery: RecoveryHint(state),
	}
}

// Label deterministically names the viewer exposure.
func Label(state TransportState, decision AdaptationDecision) string {
	return fmt.Sprintf("state=%s.layer=%d.bitrate=%dkbps", state, decision.Layer, decision.BitrateKbps)
}
