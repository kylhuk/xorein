package ui

import (
	"testing"

	"github.com/aether/code_aether/pkg/v13/chat"
	"github.com/aether/code_aether/pkg/v13/joinpolicy"
)

func TestPanelLabelForMember(t *testing.T) {
	label := Panel{SpaceName: "Ops", Mode: joinpolicy.ModeOpen, Member: true}.Label()
	if label != "Ops • Joined (open)" {
		t.Fatalf("unexpected label %q", label)
	}
}

func TestComposerStateHasDraft(t *testing.T) {
	cs := ComposerState{Draft: "hi"}
	if !cs.HasDraft() {
		t.Fatalf("expected draft to be detected")
	}
}

func TestDeliveryLabelIncludesState(t *testing.T) {
	msg := chat.Message{ChannelID: "chan", Sender: "alice", State: chat.DeliveryStateDelivered}
	if DeliveryLabel(msg) != "chan/alice/delivered" {
		t.Fatalf("unexpected label")
	}
}
