package v13

import (
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v13/chat"
)

func TestPerfChatDeliverySteps(t *testing.T) {
	msg, err := chat.NewMessage("perf-1", "channel-perf", "system", "perf check", time.Now())
	if err != nil {
		t.Fatalf("message creation failed: %v", err)
	}
	if msg.State != chat.DeliveryStatePending {
		t.Fatalf("perf message should start pending")
	}
	if err := msg.Fail("timeout"); err != nil {
		t.Fatalf("expected fail transition, got %v", err)
	}
	if msg.State != chat.DeliveryStateFailed {
		t.Fatalf("perf message should end failed")
	}
}
