package v13

import (
	"testing"

	"github.com/aether/code_aether/pkg/v11/relaypolicy"
	"github.com/aether/code_aether/pkg/v13/joinpolicy"
	"github.com/aether/code_aether/pkg/v13/spaces"
)

func TestJoinPolicyInviteOnlyRequiresToken(t *testing.T) {
	space, err := spaces.NewSpace("jspace", "Ops", "founder", "")
	if err != nil {
		t.Fatalf("space creation failed: %v", err)
	}
	if err := joinpolicy.ValidateRequest(space.Policy, "alice", ""); err != joinpolicy.ErrInviteTokenMissing {
		t.Fatalf("expected invite token requirement, got %v", err)
	}
}

func TestRelayNoDataPolicy(t *testing.T) {
	if err := relaypolicy.ValidateMode(relaypolicy.PersistenceModeDurableMessageBody); err == nil {
		t.Fatalf("relay policy should block durable message body storage")
	}
}
