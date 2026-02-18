package enforcement

import (
	"fmt"
	"sort"

	"github.com/aether/code_aether/pkg/v16/acl"
	"github.com/aether/code_aether/pkg/v16/rbac"
)

// EnforcementEngine wires RBAC and ACL evaluations for user actions.
type EnforcementEngine struct {
	RBAC        *rbac.RoleManager
	ChannelACLs map[string]*acl.ACL
}

// New creates an enforcement engine tied to the provided RBAC manager.
func New(rbacManager *rbac.RoleManager) *EnforcementEngine {
	return &EnforcementEngine{
		RBAC:        rbacManager,
		ChannelACLs: make(map[string]*acl.ACL),
	}
}

// SetChannelPolicy registers a channel-specific ACL overlay.
func (e *EnforcementEngine) SetChannelPolicy(channel string, policy *acl.ACL) {
	if channel == "" {
		return
	}
	e.ChannelACLs[channel] = policy
}

// Ensure evaluates whether the user can perform action in the channel.
func (e *EnforcementEngine) Ensure(action acl.Action, user, channel string) acl.ACLDecision {
	base := acl.New()
	perm := permissionForAction(action)
	if perm == "" {
		base.Deny(action, "unsupported action")
		return base.Evaluate(action)
	}
	if e.RBAC.WithPermission(user, perm) {
		base.Allow(action, "role-permission")
	} else {
		base.Deny(action, "missing-permission")
	}
	if channel != "" {
		if overlay, ok := e.ChannelACLs[channel]; ok {
			merged := base.Merge(overlay)
			return merged.Evaluate(action)
		}
	}
	return base.Evaluate(action)
}

// EnsureAdminAction checks whether the user may run admin workflows.
func (e *EnforcementEngine) EnsureAdminAction(user string) acl.ACLDecision {
	base := acl.New()
	if e.RBAC.WithPermission(user, rbac.PermissionAdminConfig) {
		base.Allow(acl.ActionAdminManage, "admin-role")
		return base.Evaluate(acl.ActionAdminManage)
	}
	base.Deny(acl.ActionAdminManage, "admin-only")
	return base.Evaluate(acl.ActionAdminManage)
}

func permissionForAction(action acl.Action) rbac.Permission {
	switch action {
	case acl.ActionChatSend:
		return rbac.PermissionChatSend
	case acl.ActionVoiceJoin:
		return rbac.PermissionVoiceJoin
	case acl.ActionScreenshareView:
		return rbac.PermissionScreenshareView
	default:
		return ""
	}
}

// ExplainChannelPolicy returns the channel policy trace for diagnostics.
func (e *EnforcementEngine) ExplainChannelPolicy(channel string) ([]string, error) {
	policy, ok := e.ChannelACLs[channel]
	if !ok || policy == nil {
		return nil, fmt.Errorf("no policy for %s", channel)
	}
	trace := make([]string, 0, len(policy.AllowEntries())+len(policy.DenyEntries()))
	for action, src := range policy.AllowEntries() {
		trace = append(trace, fmt.Sprintf("allow %s via %s", action, src))
	}
	for action, src := range policy.DenyEntries() {
		trace = append(trace, fmt.Sprintf("deny %s via %s", action, src))
	}
	sort.Strings(trace)
	return trace, nil
}
