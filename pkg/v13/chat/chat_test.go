package chat

import (
	"testing"
	"time"
)

func TestNewMessageRequiresFields(t *testing.T) {
	if _, err := NewMessage("", "channel", "alice", "hello", time.Now()); err == nil {
		t.Fatalf("expected error for blank id")
	}
}

func TestAcknowledgeTransitions(t *testing.T) {
	msg, _ := NewMessage("msg-1", "chan", "alice", "hi", time.Now())
	if err := msg.Acknowledge(); err != nil {
		t.Fatalf("unexpected ack error: %v", err)
	}
	if msg.State != DeliveryStateDelivered {
		t.Fatalf("expected delivered, got %q", msg.State)
	}
}

func TestFailStateRecordsReason(t *testing.T) {
	msg, _ := NewMessage("msg-2", "chan", "alice", "hi", time.Now())
	msg.Fail("network")
	if msg.State != DeliveryStateFailed {
		t.Fatalf("expected failed state")
	}
	if msg.Body == "hi" {
		t.Fatalf("expected reason appended to body")
	}
}

func TestReadMarkerTouch(t *testing.T) {
	marker := ReadMarker{ChannelID: "chan-1"}
	now := time.Now()
	if err := marker.Touch("msg-3", "bob", now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if marker.LatestMessage != "msg-3" {
		t.Fatalf("latest message not recorded")
	}
}
