package ui

import (
	"testing"

	"github.com/aether/code_aether/pkg/v15/capture"
	"github.com/aether/code_aether/pkg/v15/screenshare"
)

func TestComposeControlState(t *testing.T) {
	state := ComposeControlState(capture.SourceDisplay, screenshare.StateActive, 3200)
	if !state.CanStop || state.CanStart {
		t.Fatalf("unexpected control flags %v", state)
	}
	if state.QualityIndicator != "q=2500kbps" {
		t.Fatalf("unexpected quality indicator %s", state.QualityIndicator)
	}
	fallback := ComposeControlState(capture.SourceWindow, screenshare.StateFallback, 900)
	if fallback.Hint == "" || fallback.AdaptationLayer != 0 {
		t.Fatalf("unexpected fallback state %v", fallback)
	}
}
