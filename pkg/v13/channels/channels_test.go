package channels

import (
	"testing"

	"github.com/aether/code_aether/pkg/v13/spaces"
)

func TestNewChannelBinding(t *testing.T) {
	space, _ := spaces.NewSpace("space-1", "Team", "founder", "")
	ch, err := NewChannel("chan-1", space.ID, "general")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ch.BelongsToSpace(space.ID) {
		t.Fatalf("channel should belong to space")
	}
}

func TestChannelValidateRejectsMissingSpaceMember(t *testing.T) {
	space, _ := spaces.NewSpace("space-2", "Team", "founder", "")
	ch, _ := NewChannel("chan-2", space.ID, "random")
	ch.AddMember("ghost")
	if err := ch.Validate(space); err == nil {
		t.Fatalf("expected error for out-of-space member")
	}
}
