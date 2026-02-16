package automod

import (
	"testing"

	"github.com/aether/code_aether/pkg/v04/securitymode"
)

func TestTriggerPrecedence(t *testing.T) {
	tests := []struct {
		name    string
		trigger TriggerType
		want    int
	}{
		{name: "rate", trigger: TriggerRate, want: 3},
		{name: "keyword", trigger: TriggerKeyword, want: 2},
		{name: "extensible", trigger: TriggerExtensible, want: 1},
		{name: "unknown", trigger: TriggerType("custom"), want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TriggerPrecedence(tt.trigger); got != tt.want {
				t.Fatalf("TriggerPrecedence(%s): got %d want %d", tt.trigger, got, tt.want)
			}
		})
	}
}

func TestActionFor(t *testing.T) {
	tests := []struct {
		name    string
		trigger TriggerType
		mode    securitymode.ChannelMode
		want    ActionType
	}{
		{name: "rate clear", trigger: TriggerRate, mode: securitymode.ModeClear, want: ActionWarn},
		{name: "rate tree", trigger: TriggerRate, mode: securitymode.ModeTree, want: ActionQuarantine},
		{name: "keyword e2ee", trigger: TriggerKeyword, mode: securitymode.ModeE2EE, want: ActionEscalate},
		{name: "keyword clear", trigger: TriggerKeyword, mode: securitymode.ModeClear, want: ActionBlock},
		{name: "extensible", trigger: TriggerExtensible, mode: securitymode.ModeCrowd, want: ActionEscalate},
		{name: "unknown", trigger: TriggerType("other"), mode: securitymode.ModeChannel, want: ActionWarn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ActionFor(tt.trigger, tt.mode); got != tt.want {
				t.Fatalf("ActionFor(%s, %s): got %s want %s", tt.trigger, tt.mode, got, tt.want)
			}
		})
	}
}

func TestBypassPermitted(t *testing.T) {
	tests := []struct {
		name   string
		reason BypassReason
		mode   securitymode.ChannelMode
		want   bool
	}{
		{name: "manual clear", reason: BypassManualOverride, mode: securitymode.ModeClear, want: true},
		{name: "manual e2ee", reason: BypassManualOverride, mode: securitymode.ModeE2EE, want: false},
		{name: "policy crowd", reason: BypassPolicyExempt, mode: securitymode.ModeCrowd, want: true},
		{name: "policy channel", reason: BypassPolicyExempt, mode: securitymode.ModeChannel, want: false},
		{name: "evidence tree", reason: BypassEvidenceLag, mode: securitymode.ModeTree, want: true},
		{name: "evidence e2ee", reason: BypassEvidenceLag, mode: securitymode.ModeE2EE, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BypassPermitted(tt.reason, tt.mode); got != tt.want {
				t.Fatalf("BypassPermitted(%s, %s): got %t want %t", tt.reason, tt.mode, got, tt.want)
			}
		})
	}
}

func TestFailureMessage(t *testing.T) {
	tests := []struct {
		name   string
		reason FailureReason
		mode   securitymode.ChannelMode
		want   string
	}{
		{name: "policy missing", reason: FailurePolicyMissing, mode: securitymode.ModeClear, want: "policy version unavailable"},
		{name: "evidence error", reason: FailureEvidenceError, mode: securitymode.ModeTree, want: "audit evidence incomplete for " + string(securitymode.ModeTree)},
		{name: "mode mismatch", reason: FailureModeMismatch, mode: securitymode.ModeChannel, want: "non-compliant security mode for automation"},
		{name: "unknown", reason: FailureReason("custom"), mode: securitymode.ModeClear, want: "unknown failure"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FailureMessage(tt.reason, tt.mode); got != tt.want {
				t.Fatalf("FailureMessage(%s): got %q want %q", tt.reason, got, tt.want)
			}
		})
	}
}

func TestRequiresEvidence(t *testing.T) {
	tests := []struct {
		name    string
		trigger TriggerType
		mode    securitymode.ChannelMode
		want    bool
	}{
		{name: "e2ee rate", trigger: TriggerRate, mode: securitymode.ModeE2EE, want: false},
		{name: "e2ee keyword", trigger: TriggerKeyword, mode: securitymode.ModeE2EE, want: true},
		{name: "clear extensible", trigger: TriggerExtensible, mode: securitymode.ModeClear, want: false},
		{name: "clear rate", trigger: TriggerRate, mode: securitymode.ModeClear, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RequiresEvidence(tt.trigger, tt.mode); got != tt.want {
				t.Fatalf("RequiresEvidence(%s, %s): got %t want %t", tt.trigger, tt.mode, got, tt.want)
			}
		})
	}
}
