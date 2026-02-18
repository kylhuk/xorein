package v15

import (
	"testing"

	"github.com/aether/code_aether/pkg/v15/capture"
	"github.com/aether/code_aether/pkg/v15/screenshare"
)

func TestFirstFrameTransition(t *testing.T) {
	source, err := capture.ParseSource("display")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if source != capture.SourceDisplay {
		t.Fatalf("expected display source, got %s", source)
	}
	state := screenshare.Transition(screenshare.StateIdle, screenshare.EventStart)
	if state != screenshare.StateConnecting {
		t.Fatalf("expected connecting, got %s", state)
	}
	state = screenshare.Transition(state, screenshare.EventConnected)
	if state != screenshare.StateActive {
		t.Fatalf("expected active, got %s", state)
	}
}
