package v22

import (
	"testing"

	"github.com/aether/code_aether/pkg/v22/history/apply"
)

func TestE2ERedactionBackfillFlow(t *testing.T) {
	applier := apply.NewApplier()
	applier.ApplyTombstones([]apply.Tombstone{{MessageID: "m", Reason: "policy"}})
	applier.ApplyEvents([]apply.BackfillEvent{{MessageID: "m", Content: "secret"}})
	if len(applier.VisibleEvents()) != 0 {
		t.Fatalf("expected tombstoned events to be hidden")
	}
}
