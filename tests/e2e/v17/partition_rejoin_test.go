package v17_test

import (
	"testing"

	"github.com/aether/code_aether/pkg/v17/moderation"
	"github.com/aether/code_aether/pkg/v17/modsync"
)

func TestPartitionRejoinOrdering(t *testing.T) {
	orderer := modsync.NewOrderer()
	partitionA := []moderation.SignedEvent{
		{ID: "evt-part-1", Room: "room:delta", Actor: "mod@relay", Target: "user:alpha", Type: moderation.EventSlowMode, Timestamp: 1, Signature: "sig:mod@relay"},
	}
	partitionB := []moderation.SignedEvent{
		{ID: "evt-part-2", Room: "room:delta", Actor: "mod@relay", Target: "user:beta", Type: moderation.EventBan, Timestamp: 2, Signature: "sig:mod@relay"},
		partitionA[0],
	}
	merged := orderer.Merge(partitionA, partitionB)
	if len(merged) != 2 {
		t.Fatalf("expected merged len 2, got %d", len(merged))
	}
	if merged[0].Timestamp != 1 || merged[1].Timestamp != 2 {
		t.Fatalf("expected chronological ordering, got %v", []int{int(merged[0].Timestamp), int(merged[1].Timestamp)})
	}
}
