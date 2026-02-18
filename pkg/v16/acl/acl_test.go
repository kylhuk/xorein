package acl

import "testing"

func TestACLMergePrecedence(t *testing.T) {
	base := New()
	base.Allow(ActionChatSend, "channel")
	overlay := New()
	overlay.Deny(ActionChatSend, "admin")
	merged := base.Merge(overlay)
	decision := merged.Evaluate(ActionChatSend)
	if decision.Allowed {
		t.Fatalf("expected deny to override allow")
	}
	if decision.Reason == "" {
		t.Fatalf("reason expected")
	}
}

func TestExplainabilityTrace(t *testing.T) {
	acl := New()
	acl.Allow(ActionVoiceJoin, "policy")
	acl.Allow(ActionScreenshareView, "policy")
	acl.Deny(ActionScreenshareView, "override")
	dec := acl.Evaluate(ActionScreenshareView)
	if dec.Allowed {
		t.Fatalf("should be denied")
	}
	if len(dec.Trace) != 1 {
		t.Fatalf("expected one trace entry, got %v", dec.Trace)
	}
}
