package capture

import (
	"fmt"
	"strings"
)

// SourceType represents a deterministic capture source category.
type SourceType string

const (
	SourceDisplay SourceType = "display"
	SourceWindow  SourceType = "window"
)

// ErrorClass categorizes capture failures for deterministic handling.
type ErrorClass string

const (
	ErrorClassSourceSelection ErrorClass = "capture-source-selection"
	ErrorClassPresetMapping   ErrorClass = "preset-mapping"
)

// CaptureError wraps deterministic error information.
type CaptureError struct {
	Class   ErrorClass
	Message string
}

func (e CaptureError) Error() string {
	return fmt.Sprintf("%s: %s", e.Class, e.Message)
}

// Preset defines a quality tier for screen share capture.
type Preset string

const (
	PresetHigh   Preset = "high"
	PresetMedium Preset = "medium"
	PresetLow    Preset = "low"
)

// EncoderProfile binds an encoder name to a target bitrate (kbps).
type EncoderProfile struct {
	Encoder     string
	BitrateKbps int
}

var presetProfiles = map[Preset]EncoderProfile{
	PresetHigh: {
		Encoder:     "vp9",
		BitrateKbps: 4500,
	},
	PresetMedium: {
		Encoder:     "h264",
		BitrateKbps: 2500,
	},
	PresetLow: {
		Encoder:     "av1",
		BitrateKbps: 1200,
	},
}

// ParseSource converts an input string into a deterministic SourceType.
func ParseSource(raw string) (SourceType, error) {
	normalized := SourceType(strings.ToLower(strings.TrimSpace(raw)))
	switch normalized {
	case "", SourceDisplay:
		return SourceDisplay, nil
	case SourceWindow:
		return SourceWindow, nil
	default:
		return "", CaptureError{Class: ErrorClassSourceSelection, Message: fmt.Sprintf("unknown source %q", raw)}
	}
}

// ProfileForPreset returns the encoder profile associated with a preset tier.
func ProfileForPreset(preset Preset) (EncoderProfile, error) {
	profile, ok := presetProfiles[preset]
	if !ok {
		return EncoderProfile{}, CaptureError{Class: ErrorClassPresetMapping, Message: fmt.Sprintf("unsupported preset %q", preset)}
	}
	return profile, nil
}
