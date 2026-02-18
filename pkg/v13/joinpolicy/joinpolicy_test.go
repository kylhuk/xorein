package joinpolicy

import "testing"

func TestParseModeDefaults(t *testing.T) {
	mode, err := ParseMode("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != ModeInviteOnly {
		t.Fatalf("expected default invite-only, got %q", mode)
	}
}

func TestParseModeRejectsUnknown(t *testing.T) {
	if _, err := ParseMode("unknown"); err == nil {
		t.Fatalf("expected error for unknown mode")
	}
}

func TestValidateRequestInviteOnly(t *testing.T) {
	if err := ValidateRequest(ModeInviteOnly, "alice", ""); err != ErrInviteTokenMissing {
		t.Fatalf("expected invite token missing, got %v", err)
	}
}

func TestValidateRequestRequestToJoin(t *testing.T) {
	if err := ValidateRequest(ModeRequestToJoin, "bob", ""); err != nil {
		t.Fatalf("expected request to join allowed without token, got %v", err)
	}
}
