package v17perf_test

import (
	"fmt"
	"testing"

	"github.com/aether/code_aether/pkg/v17/moderation"
)

func TestModerationStepsPerformance(t *testing.T) {
	engine := moderation.NewEngine()
	room := "room:perf"
	for i := int64(1); i <= 5; i++ {
		evt := moderation.SignedEvent{
			ID:        fmt.Sprintf("evt-perf-%d", i),
			Room:      room,
			Actor:     "mod@relay",
			Target:    fmt.Sprintf("user:%d", i),
			Type:      moderation.EventSlowMode,
			Timestamp: i * 10,
			Signature: "sig:mod@relay",
		}
		res := engine.Apply(evt)
		if !res.Accepted {
			t.Fatalf("step %d rejected: %s", i, res.Reason)
		}
	}
	if ts := engine.LastTimestamp(room); ts != 50 {
		t.Fatalf("expected final timestamp 50, got %d", ts)
	}
}
