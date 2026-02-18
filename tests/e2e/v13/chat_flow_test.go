package v13

import (
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v13/channels"
	"github.com/aether/code_aether/pkg/v13/chat"
	"github.com/aether/code_aether/pkg/v13/spaces"
)

func TestChatFlowAcknowledge(t *testing.T) {
	space, _ := spaces.NewSpace("cspace", "Chat", "founder", "")
	ch, err := channels.NewChannel("c1", space.ID, "general")
	if err != nil {
		t.Fatalf("channel init failed: %v", err)
	}
	ch.AddMember("founder")
	msg, err := chat.NewMessage("m1", ch.ID, "founder", "hello", time.Now())
	if err != nil {
		t.Fatalf("message creation failed: %v", err)
	}
	if err := msg.Acknowledge(); err != nil {
		t.Fatalf("ack failed: %v", err)
	}
	if msg.State != chat.DeliveryStateDelivered {
		t.Fatalf("message not delivered")
	}
	marker := chat.ReadMarker{ChannelID: ch.ID}
	if err := marker.Touch(msg.ID, "founder", time.Now()); err != nil {
		t.Fatalf("read marker failed: %v", err)
	}
}
