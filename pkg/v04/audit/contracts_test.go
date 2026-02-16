package audit

import (
	"testing"

	"github.com/aether/code_aether/pkg/v04/policy"
	"github.com/aether/code_aether/pkg/v04/roles"
	"github.com/aether/code_aether/pkg/v04/securitymode"
)

func TestAuditEntryTraceKeyAndSignedTrace(t *testing.T) {
	entry := AuditEntry{
		Event:  EventAutoMod,
		Policy: policy.PolicyVersion{ID: "policy", Major: 1, Minor: 0},
		Role:   roles.RoleAdmin,
		Mode:   securitymode.ModeClear,
	}

	wantKey := "admin|auto_mod|1.0:policy|Clear"
	if got := entry.TraceKey(); got != wantKey {
		t.Fatalf("TraceKey: got %s want %s", got, wantKey)
	}

	if got := entry.SignedTrace(); got != "unsigned" {
		t.Fatalf("SignedTrace unsigned: got %s", got)
	}

	entry.Signed = true
	if got := entry.SignedTrace(); got != "signed:"+wantKey {
		t.Fatalf("SignedTrace signed: got %s", got)
	}
}

func TestVisibilityFor(t *testing.T) {
	tests := []struct {
		name string
		role roles.Role
		mode securitymode.ChannelMode
		want bool
	}{
		{name: "owner always", role: roles.RoleOwner, mode: securitymode.ModeTree, want: true},
		{name: "moderator e2ee", role: roles.RoleModerator, mode: securitymode.ModeE2EE, want: false},
		{name: "moderator clear", role: roles.RoleModerator, mode: securitymode.ModeClear, want: true},
		{name: "member clear", role: roles.RoleMember, mode: securitymode.ModeClear, want: true},
		{name: "member channel", role: roles.RoleMember, mode: securitymode.ModeChannel, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VisibilityFor(tt.role, tt.mode); got != tt.want {
				t.Fatalf("VisibilityFor(%s, %s): got %t want %t", tt.role, tt.mode, got, tt.want)
			}
		})
	}
}

func TestDisclosureRequired(t *testing.T) {
	if !DisclosureRequired(securitymode.ModeTree) {
		t.Fatal("DisclosureRequired should be true for known mode")
	}

	if DisclosureRequired(securitymode.ChannelMode("unknown")) {
		t.Fatal("DisclosureRequired should be false for unknown mode")
	}
}
