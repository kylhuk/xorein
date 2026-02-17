package discovery

import (
	"reflect"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

func TestDiscoveryContractHelpers(t *testing.T) {
	tests := []struct {
		name       string
		contract   DiscoveryContract
		wantAnchor string
		wantReason string
	}{
		{
			name:       "freshness",
			contract:   NewDiscoveryContract("S6-01", "T6-01", "artifact", conformance.GateV6G0, DiscoveryReasonFreshness),
			wantAnchor: "artifact#T6-01",
			wantReason: string(DiscoveryReasonFreshness),
		},
		{
			name:       "consistency",
			contract:   NewDiscoveryContract("S6-02", "T6-02", "doc", conformance.GateV6G1, DiscoveryReasonConsistency),
			wantAnchor: "doc#T6-02",
			wantReason: string(DiscoveryReasonConsistency),
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

func TestDiscoveryReasonClassesDeterministic(t *testing.T) {
	want := []DiscoveryReasonClass{DiscoveryReasonFreshness, DiscoveryReasonPoisoning, DiscoveryReasonConsistency}
	got := DiscoveryReasonClasses()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reason classes changed: want %v got %v", want, got)
	}
}

func TestAssessFreshnessStates(t *testing.T) {
	tests := []struct {
		name        string
		entryAge    time.Duration
		ttl         time.Duration
		attempts    int
		maxAttempts int
		wantStatus  FreshnessStatus
		wantReason  string
	}{
		{
			name:        "valid",
			entryAge:    2 * time.Minute,
			ttl:         5 * time.Minute,
			attempts:    0,
			maxAttempts: 3,
			wantStatus:  FreshnessStatusValid,
			wantReason:  "discovery.freshness.success",
		},
		{
			name:        "retry",
			entryAge:    15 * time.Minute,
			ttl:         10 * time.Minute,
			attempts:    1,
			maxAttempts: 3,
			wantStatus:  FreshnessStatusRetry,
			wantReason:  "discovery.freshness.retry",
		},
		{
			name:        "stale",
			entryAge:    20 * time.Minute,
			ttl:         10 * time.Minute,
			attempts:    3,
			maxAttempts: 3,
			wantStatus:  FreshnessStatusStale,
			wantReason:  "discovery.freshness.stale",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := AssessFreshness(tt.entryAge, tt.ttl, tt.attempts, tt.maxAttempts); got.Status != tt.wantStatus || got.ReasonLabel != tt.wantReason {
				t.Fatalf("freshness mismatch: want status %s reason %q got %s %q", tt.wantStatus, tt.wantReason, got.Status, got.ReasonLabel)
			}
		})
	}
}

func TestRetryBackoffBounds(t *testing.T) {
	base := 2 * time.Second
	tests := []struct {
		name    string
		attempt int
		want    time.Duration
	}{
		{name: "zero", attempt: 0, want: base},
		{name: "multiply", attempt: 3, want: 8 * base},
		{name: "capped", attempt: 10, want: 30 * time.Second},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := RetryBackoff(tt.attempt, base); got != tt.want {
				t.Fatalf("backoff mismatch: want %v got %v", tt.want, got)
			}
		})
	}
}

func TestClassifyPoisonAttempt(t *testing.T) {
	tests := []struct {
		name             string
		validSignature   bool
		conflictingPeers int
		wantClass        string
		wantReason       string
	}{
		{name: "clean", validSignature: true, wantClass: "clean", wantReason: "discovery.poison.clean"},
		{name: "detected", conflictingPeers: 0, wantClass: "detected", wantReason: "discovery.poison.detected"},
		{name: "invalid", conflictingPeers: 2, wantClass: "invalid", wantReason: "discovery.poison.invalid"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyPoisonAttempt(tt.validSignature, tt.conflictingPeers)
			if got.Classification != tt.wantClass || got.ReasonLabel != tt.wantReason {
				t.Fatalf("classification mismatch: want %s/%s got %s/%s", tt.wantClass, tt.wantReason, got.Classification, got.ReasonLabel)
			}
		})
	}
}

func TestResolveMultiSourceConflictOrdering(t *testing.T) {
	tests := []struct {
		name       string
		sources    map[string]int64
		wantCanon  string
		wantCount  int
		wantReason string
	}{
		{name: "empty", sources: nil, wantCanon: "", wantCount: 0, wantReason: "discovery.consistency.conflict"},
		{name: "single", sources: map[string]int64{"a": 10}, wantCanon: "a", wantCount: 0, wantReason: "discovery.consistency.success"},
		{name: "conflict", sources: map[string]int64{"a": 5, "b": 5, "c": 4}, wantCanon: "a", wantCount: 2, wantReason: "discovery.consistency.conflict"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveMultiSourceConflict(tt.sources)
			if got.CanonicalSource != tt.wantCanon || got.ConflictCount != tt.wantCount || got.ReasonLabel != tt.wantReason {
				t.Fatalf("conflict mismatch: want %s/%d/%s got %s/%d/%s", tt.wantCanon, tt.wantCount, tt.wantReason, got.CanonicalSource, got.ConflictCount, got.ReasonLabel)
			}
		})
	}
}
