package v15

import (
	"testing"

	"github.com/aether/code_aether/pkg/v11/relaypolicy"
	"github.com/aether/code_aether/pkg/v15/screenshare"
)

func TestAdaptationRecovery(t *testing.T) {
	state := screenshare.Transition(screenshare.StateActive, screenshare.EventQualityDrop)
	if state != screenshare.StateFallback {
		t.Fatalf("expected fallback, got %s", state)
	}
	hint := screenshare.RecoveryHint(state)
	if hint == "" {
		t.Fatal("expected recovery hint")
	}
}

func TestRelayNoDataRegression(t *testing.T) {
	forbidden := relaypolicy.ForbiddenClasses()
	if len(forbidden) == 0 {
		t.Fatal("expected forbidden classes")
	}
	containsArchive := false
	for _, cls := range forbidden {
		if cls == relaypolicy.StorageClassMediaFrameArchive {
			containsArchive = true
		}
	}
	if !containsArchive {
		t.Fatalf("media frame archive should remain forbidden: %+v", forbidden)
	}
}
