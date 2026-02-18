package acl

import (
	"fmt"
	"sort"
)

// Action identifies permission actions evaluated via ACLs.
type Action string

const (
	ActionChatSend        Action = "chat:send"
	ActionVoiceJoin       Action = "voice:join"
	ActionScreenshareView Action = "screenshare:view"
	ActionAdminManage     Action = "admin:manage"
)

// ACLDecision captures the outcome and rationale for an action.
type ACLDecision struct {
	Allowed bool
	Reason  string
	Trace   []string
}

// ACL stores allow/deny rules with explainable sources.
type ACL struct {
	allow map[Action]string
	deny  map[Action]string
}

// New instantiates an empty ACL.
func New() *ACL {
	return &ACL{
		allow: make(map[Action]string),
		deny:  make(map[Action]string),
	}
}

// Allow records an allow rule for an action.
func (a *ACL) Allow(action Action, source string) {
	if a == nil {
		return
	}
	if source == "" {
		source = "allow"
	}
	if _, denied := a.deny[action]; denied {
		return
	}
	a.allow[action] = fmt.Sprintf("%s", source)
}

// Deny records a deny rule for an action with reason.
func (a *ACL) Deny(action Action, source string) {
	if a == nil {
		return
	}
	if source == "" {
		source = "deny"
	}
	a.deny[action] = fmt.Sprintf("%s", source)
	delete(a.allow, action)
}

// Merge deterministically combines base ACLs with overrides.
func (a *ACL) Merge(overlay *ACL) *ACL {
	merged := New()
	if a != nil {
		for act, src := range a.allow {
			merged.allow[act] = fmt.Sprintf("base:%s", src)
		}
		for act, src := range a.deny {
			merged.deny[act] = fmt.Sprintf("base:%s", src)
		}
	}
	if overlay != nil {
		for act, src := range overlay.allow {
			if _, denied := merged.deny[act]; denied {
				continue
			}
			merged.allow[act] = fmt.Sprintf("overlay:%s", src)
		}
		for act, src := range overlay.deny {
			merged.deny[act] = fmt.Sprintf("overlay:%s", src)
			delete(merged.allow, act)
		}
	}
	return merged
}

// Evaluate returns the decision for the requested action.
func (a *ACL) Evaluate(action Action) ACLDecision {
	if a == nil {
		return ACLDecision{Allowed: false, Reason: "no policy"}
	}
	if reason, denied := a.deny[action]; denied {
		return ACLDecision{Allowed: false, Reason: fmt.Sprintf("denied: %s", reason), Trace: a.trace(action)}
	}
	if reason, ok := a.allow[action]; ok {
		return ACLDecision{Allowed: true, Reason: fmt.Sprintf("allowed: %s", reason), Trace: a.trace(action)}
	}
	return ACLDecision{Allowed: false, Reason: "no rule", Trace: a.trace(action)}
}

func (a *ACL) trace(action Action) []string {
	entries := make([]string, 0, 2)
	if reason, ok := a.allow[action]; ok {
		entries = append(entries, fmt.Sprintf("allow:%s", reason))
	}
	if reason, ok := a.deny[action]; ok {
		entries = append(entries, fmt.Sprintf("deny:%s", reason))
	}
	sort.Strings(entries)
	return entries
}

// AllowEntries returns a snapshot of allow rules.
func (a *ACL) AllowEntries() map[Action]string {
    entries := make(map[Action]string, len(a.allow))
    for action, reason := range a.allow {
        entries[action] = reason
    }
    return entries
}

// DenyEntries returns a snapshot of deny rules.
func (a *ACL) DenyEntries() map[Action]string {
    entries := make(map[Action]string, len(a.deny))
    for action, reason := range a.deny {
        entries[action] = reason
    }
    return entries
}
