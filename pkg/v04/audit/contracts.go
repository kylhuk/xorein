package audit

import (
	"fmt"
	"strings"

	"github.com/aether/code_aether/pkg/v04/policy"
	"github.com/aether/code_aether/pkg/v04/roles"
	"github.com/aether/code_aether/pkg/v04/securitymode"
)

type EventType string

const (
	EventPolicyVersion EventType = "policy_version"
	EventModeChange    EventType = "mode_change"
	EventAutoMod       EventType = "auto_mod"
)

type AuditEntry struct {
	Event   EventType
	Policy  policy.PolicyVersion
	Role    roles.Role
	Mode    securitymode.ChannelMode
	Signed  bool
	Details string
}

func (e AuditEntry) TraceKey() string {
	parts := []string{string(e.Role), string(e.Event), e.Policy.String(), string(e.Mode)}
	return strings.Join(parts, "|")
}

func (e AuditEntry) SignedTrace() string {
	if !e.Signed {
		return "unsigned"
	}
	return fmt.Sprintf("signed:%s", e.TraceKey())
}

func VisibilityFor(role roles.Role, mode securitymode.ChannelMode) bool {
	switch role {
	case roles.RoleOwner, roles.RoleAdmin:
		return true
	case roles.RoleModerator:
		return mode != securitymode.ModeE2EE
	default:
		return mode == securitymode.ModeClear
	}
}

func DisclosureRequired(mode securitymode.ChannelMode) bool {
	requirement := securitymode.DisclosureFor(mode)
	return requirement.Description != "Unknown mode"
}
