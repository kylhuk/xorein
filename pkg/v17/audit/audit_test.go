package audit_test

import (
	"testing"

	"github.com/aether/code_aether/pkg/v17/audit"
	"github.com/aether/code_aether/pkg/v17/moderation"
)

func TestAuditLogVisibility(t *testing.T) {
	log := audit.NewAuditLog()
	event := moderation.SignedEvent{ID: "evt-log", Room: "room:logs", Actor: "mod@relay", Target: "user:xyz", Type: moderation.EventSlowMode, Timestamp: 999, Signature: "sig:mod@relay"}
	log.Append(event, []audit.Role{audit.RoleModerator})
	if got := log.Query(audit.RoleObserver); len(got) != 0 {
		t.Fatalf("observer should not see moderator-only entry, got %d rows", len(got))
	}
	if got := log.Query(audit.RoleModerator); len(got) != 1 {
		t.Fatalf("moderator should see entry, got %d", len(got))
	}
	if got := log.Query(audit.RoleAdmin); len(got) != 0 {
		t.Fatalf("admin should not see nonexistent entry, got %d", len(got))
	}
}

func TestAuditEntriesCopy(t *testing.T) {
	log := audit.NewAuditLog()
	entry := log.Append(moderation.SignedEvent{ID: "evt-copy", Room: "room:copy", Actor: "mod@relay", Target: "user:test", Type: moderation.EventBan, Timestamp: 100, Signature: "sig:mod@relay"}, nil)
	entries := log.Entries()
	if entries[0].Sequence != entry.Sequence {
		t.Fatalf("expected sequence %d, got %d", entry.Sequence, entries[0].Sequence)
	}
	entries[0].Sequence = -1
	if log.Entries()[0].Sequence != entry.Sequence {
		t.Fatalf("entries should be append-only, mutated copy leaked")
	}
}
