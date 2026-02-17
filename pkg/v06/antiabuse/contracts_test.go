package antiabuse

import (
	"reflect"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

func TestAntiAbuseContractHelpers(t *testing.T) {
	tests := []struct {
		name       string
		contract   AntiAbuseContract
		wantAnchor string
		wantReason string
	}{
		{
			name:       "pow",
			contract:   NewAntiAbuseContract("pow-policy", "T6-20", "pow-doc", conformance.GateV6G2, AntiAbuseReasonPoW),
			wantAnchor: "pow-doc#T6-20",
			wantReason: string(AntiAbuseReasonPoW),
		},
		{
			name:       "score",
			contract:   NewAntiAbuseContract("score-policy", "T6-21", "score-doc", conformance.GateV6G2, AntiAbuseReasonScore),
			wantAnchor: "score-doc#T6-21",
			wantReason: string(AntiAbuseReasonScore),
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

func TestAntiAbuseReasonClassesDeterministic(t *testing.T) {
	want := []AntiAbuseReasonClass{AntiAbuseReasonPoW, AntiAbuseReasonLimiter, AntiAbuseReasonScore}
	got := AntiAbuseReasonClasses()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reason classes changed: want %v got %v", want, got)
	}
}

func TestPowProfileAdaptBounds(t *testing.T) {
	profile := PowProfile{EnvelopeMin: 2, EnvelopeMax: 10}
	tests := []struct {
		name    string
		desired uint
		want    uint
	}{
		{name: "below_min", desired: 1, want: 2},
		{name: "within", desired: 5, want: 5},
		{name: "above_max", desired: 20, want: 10},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := profile.Adapt(tt.desired); got != tt.want {
				t.Fatalf("adapt mismatch: want %d got %d", tt.want, got)
			}
		})
	}
}

func TestDecideLimiterThresholds(t *testing.T) {
	const threshold = 10
	const burst = 5
	tests := []struct {
		name       string
		current    int
		wantStatus string
		wantRetry  time.Duration
	}{
		{name: "ok", current: 5, wantStatus: "antiabuse.limiter.ok", wantRetry: 0},
		{name: "burst", current: 12, wantStatus: "antiabuse.limiter.throttled", wantRetry: 2 * time.Second},
		{name: "overburst", current: 20, wantStatus: "antiabuse.limiter.throttled", wantRetry: 5 * time.Second},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := DecideLimiter(tt.current, threshold, burst)
			if got.Status != tt.wantStatus || got.RetryAfter != tt.wantRetry {
				t.Fatalf("limiter mismatch: want %s/%v got %s/%v", tt.wantStatus, tt.wantRetry, got.Status, got.RetryAfter)
			}
		})
	}
}

func TestEvaluatePeerScoreReintegration(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		score       float64
		penalized   bool
		lastPenalty time.Time
		window      time.Duration
		wantReason  string
		wantReint   bool
	}{
		{name: "allow", score: 0.5, penalized: false, window: time.Second, wantReason: "antiabuse.score.allow", wantReint: false},
		{name: "penalize", score: -1.5, penalized: true, lastPenalty: now.Add(-30 * time.Second), window: time.Minute, wantReason: "antiabuse.score.penalize", wantReint: false},
		{name: "reintegrate", score: 0, penalized: true, lastPenalty: now.Add(-2 * time.Minute), window: time.Minute, wantReason: "antiabuse.score.reintegrate", wantReint: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluatePeerScore(tt.score, 0, tt.penalized, tt.lastPenalty, tt.window)
			if got.Reason != tt.wantReason || got.Reintegration != tt.wantReint {
				t.Fatalf("peer score mismatch: want %s/%t got %s/%t", tt.wantReason, tt.wantReint, got.Reason, got.Reintegration)
			}
			if got.Score < -1 || got.Score > 1 {
				t.Fatalf("score bounded: got %f", got.Score)
			}
		})
	}
}
