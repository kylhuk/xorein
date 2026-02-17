package filters

import (
	"reflect"
	"testing"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

func TestFilterContractHelpers(t *testing.T) {
	tests := []struct {
		name       string
		contract   FilterContract
		wantAnchor string
		wantReason string
	}{
		{
			name:       "process",
			contract:   NewFilterContract("S6-10", "T6-50", "filter-doc", conformance.GateV6G5, FilterReasonProcess),
			wantAnchor: "filter-doc#T6-50",
			wantReason: string(FilterReasonProcess),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.contract.EvidenceAnchor(); got != tt.wantAnchor {
				t.Fatalf("anchor mismatch: want %q got %q", tt.wantAnchor, got)
			}
			if got := tt.contract.ReasonLabel(); got != tt.wantReason {
				t.Fatalf("reason label mismatch: want %q got %q", tt.wantReason, got)
			}
		})
	}
}

func TestFilterReasonClassesDeterministic(t *testing.T) {
	want := []FilterReasonClass{FilterReasonProcess}
	got := FilterReasonClasses()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reason classes changed: want %v got %v", want, got)
	}
}

func TestDecideFilterExecution(t *testing.T) {
	tests := []struct {
		name         string
		mode         SecurityMode
		optional     bool
		wantAllow    bool
		wantReason   string
		wantRecovery string
	}{
		{name: "clear", mode: SecurityModeClear, wantAllow: true, wantReason: "filters.process.success"},
		{name: "e2ee-optional", mode: SecurityModeE2EE, optional: true, wantAllow: false, wantReason: "filters.process.blocked", wantRecovery: "filters.process.recover"},
		{name: "e2ee-required", mode: SecurityModeE2EE, wantAllow: true, wantReason: "filters.process.success"},
		{name: "unknown-optional", mode: "unknown", optional: true, wantAllow: true, wantReason: "filters.process.blocked"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := DecideFilterExecution(tt.mode, tt.optional)
			if got.Allowed != tt.wantAllow || got.Reason != tt.wantReason || got.Recovery != tt.wantRecovery {
				t.Fatalf("filter decision mismatch: want %t/%s/%s got %t/%s/%s", tt.wantAllow, tt.wantReason, tt.wantRecovery, got.Allowed, got.Reason, got.Recovery)
			}
		})
	}
}

func TestPolicyEnvelopeEvaluate(t *testing.T) {
	tests := []struct {
		name         string
		envelope     PolicyEnvelope
		input        int
		wantSeverity int
		wantAllowed  bool
		wantReason   string
	}{
		{name: "below_min", envelope: PolicyEnvelope{MinSeverity: 2, MaxSeverity: 5, Reason: "policy-env"}, input: 1, wantSeverity: 2, wantAllowed: true, wantReason: "filters.policy.accept"},
		{name: "above_max", envelope: PolicyEnvelope{MinSeverity: 2, MaxSeverity: 5, Reason: "policy-env"}, input: 10, wantSeverity: 5, wantAllowed: true, wantReason: "filters.policy.accept"},
		{name: "disallowed", envelope: PolicyEnvelope{MinSeverity: 10, MaxSeverity: 5, Reason: "policy-blocked"}, input: 3, wantSeverity: 5, wantAllowed: false, wantReason: "filters.policy.blocked"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			decision := tt.envelope.Evaluate(tt.input)
			if decision.Severity != tt.wantSeverity || decision.Allowed != tt.wantAllowed || decision.Reason != tt.wantReason {
				t.Fatalf("policy decision mismatch: want %d/%t/%s got %d/%t/%s", tt.wantSeverity, tt.wantAllowed, tt.wantReason, decision.Severity, decision.Allowed, decision.Reason)
			}
			if decision.EnvelopeName != tt.envelope.Reason {
				t.Fatalf("envelope mismatch: want %s got %s", tt.envelope.Reason, decision.EnvelopeName)
			}
		})
	}
}
