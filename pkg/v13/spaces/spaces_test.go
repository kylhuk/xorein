package spaces

import (
	"testing"

	"github.com/aether/code_aether/pkg/v13/joinpolicy"
)

func TestNewSpaceDefaults(t *testing.T) {
	space, err := NewSpace("sp1", "Ops", "founder", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if space.Policy != joinpolicy.ModeInviteOnly {
		t.Fatalf("expected invite-only policy, got %q", space.Policy)
	}
	if !space.Visible {
		t.Fatalf("space should default to visible")
	}
	if !space.IsMember("founder") {
		t.Fatalf("founder must be initial member")
	}
}

func TestAddMemberDedup(t *testing.T) {
	space, _ := NewSpace("sp2", "Ops", "founder", joinpolicy.ModeOpen)
	space.AddMember("alice")
	space.AddMember("alice")
	count := 0
	for _, member := range space.Members {
		if member == "alice" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected single alice member, got %d", count)
	}
}

func TestTransferFounderRequiresMember(t *testing.T) {
	space, _ := NewSpace("sp3", "Ops", "founder", joinpolicy.ModeOpen)
	err := space.TransferFounder("bob")
	if err == nil {
		t.Fatalf("expected error transferring to non-member")
	}
}
