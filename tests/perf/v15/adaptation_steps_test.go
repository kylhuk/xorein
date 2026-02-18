package v15

import (
	"testing"

	"github.com/aether/code_aether/pkg/v15/screenshare"
)

func TestAdaptationSteps(t *testing.T) {
	for _, bw := range []int{5000, 3000, 1500, 800} {
		desc := screenshare.DetermineAdaptation(bw)
		if desc.BitrateKbps > bw+200 {
			t.Fatalf("bitrate ahead of bandwidth for %d: %v", bw, desc)
		}
	}
}
