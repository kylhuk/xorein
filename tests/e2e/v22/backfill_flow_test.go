package v22

import (
	"testing"

	"github.com/aether/code_aether/pkg/v22/history/backfill"
)

func TestE2EBackfillFlow(t *testing.T) {
	manager := backfill.NewManager(1)
	req := backfill.BackfillRequest{SpaceID: "space", ChannelID: "chan", Range: backfill.TimeRange{Start: 0, End: 10}}
	err := manager.Backfill(req, func() ([]backfill.Segment, error) {
		return []backfill.Segment{{ID: "s"}}, nil
	}, func(backfill.Segment) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
