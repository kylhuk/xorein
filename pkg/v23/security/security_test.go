package security

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestQuotaEnforcer(t *testing.T) {
	enforcer := &QuotaEnforcer{MaxEntries: 2, MaxRetentionDays: 30}

	if err := enforcer.Enforce(2, 30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := enforcer.Enforce(3, 10)
	var refusal *RefusalError
	if !errors.As(err, &refusal) {
		t.Fatalf("expected refusal error, got %v", err)
	}
	if refusal.Reason != RefusalReasonQuotaEntriesExceeded {
		t.Fatalf("expected entries reason, got %s", refusal.Reason)
	}

	err = enforcer.Enforce(1, 31)
	if !errors.As(err, &refusal) {
		t.Fatalf("expected refusal error, got %v", err)
	}
	if refusal.Reason != RefusalReasonQuotaRetentionExceeded {
		t.Fatalf("expected retention reason, got %s", refusal.Reason)
	}
}

func TestRateLimiter(t *testing.T) {
	policy := RateLimitPolicy{Limit: 2, Window: time.Minute, MaxResponseBytes: 1024}
	limiter := NewRateLimiter(policy)
	now := time.Date(2026, 2, 19, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 2; i++ {
		if err := limiter.CheckRequest(now); err != nil {
			t.Fatalf("unexpected error at iteration %d: %v", i, err)
		}
	}

	if err := limiter.CheckRequest(now); err == nil {
		t.Fatalf("expected rate limit exceeded error")
	} else if !errors.As(err, new(*RefusalError)) || err.(*RefusalError).Reason != RefusalReasonRateLimitExceeded {
		t.Fatalf("unexpected error reason: %v", err)
	}

	later := now.Add(policy.Window)
	if err := limiter.CheckRequest(later); err != nil {
		t.Fatalf("expected window reset, got %v", err)
	}

	if err := limiter.ValidateResponseSize(512); err != nil {
		t.Fatalf("unexpected response size error: %v", err)
	}
	if err := limiter.ValidateResponseSize(2048); err == nil {
		t.Fatalf("expected response size refusal")
	} else if !errors.As(err, new(*RefusalError)) || err.(*RefusalError).Reason != RefusalReasonResponseSizeExceeded {
		t.Fatalf("unexpected response size reason: %v", err)
	}
}

func TestTelemetryAggregator(t *testing.T) {
	agg := NewTelemetryAggregator()
	labels := map[string]string{"event": "quota", "outcome": "refused"}
	agg.Record(labels)
	agg.Record(labels)

	snapshot := agg.Snapshot()
	key := "event=quota;outcome=refused;"
	if got := snapshot[key]; got != 2 {
		t.Fatalf("expected 2 counts for %s, got %d", key, got)
	}

	snapshot[key] = 0
	copied := agg.Snapshot()
	if got := copied[key]; got != 2 {
		t.Fatalf("snapshot should not be affected by mutation, got %d", got)
	}
}

func TestPrivacyConfigBlocksKeywords(t *testing.T) {
	cfg := PrivacyConfig{}
	req := BackfillRequest{SpaceID: "space", ChannelID: "chan", Query: "secret"}
	err := cfg.ValidateBackfillRequest(req)
	var refusal *RefusalError
	if !errors.As(err, &refusal) {
		t.Fatalf("expected refusal error, got %v", err)
	}
	if refusal.Reason != RefusalReasonKeywordBackfillNotAllowed {
		t.Fatalf("unexpected refusal reason %s", refusal.Reason)
	}
	if !strings.Contains(refusal.Details, "keyword-bearing backfill blocked by default") {
		t.Fatalf("unexpected refusal details %q", refusal.Details)
	}
	cfg.AllowKeywordBackfill = true
	if err := cfg.ValidateBackfillRequest(req); err != nil {
		t.Fatalf("keyword request should be allowed after opt-in, got %v", err)
	}
}

func TestCoverageLabeling(t *testing.T) {
	state := CoverageState{Gaps: []CoverageGap{{Start: 0, End: 10, Reason: "missing"}}}
	if got := LabelCoverage(state); got != CoverageLabelPartial {
		t.Fatalf("expected partial label when gaps remain, got %s", got)
	}
	complete := CoverageState{Complete: true}
	if got := LabelCoverage(complete); got != CoverageLabelFull {
		t.Fatalf("expected full label when complete, got %s", got)
	}
	unknown := CoverageState{}
	if got := LabelCoverage(unknown); got != CoverageLabelIncomplete {
		t.Fatalf("expected incomplete label for unknown state, got %s", got)
	}
}

func TestAssistedSearchGate(t *testing.T) {
	gate := AssistedSearchGate{}
	if gate.Allows("token") {
		t.Fatalf("disabled gate should refuse requests")
	}
	gate = gate.Enable("consent-1")
	if !gate.Allows("consent-1") {
		t.Fatalf("enabled gate should honor matching token")
	}
	if gate.Allows("other") {
		t.Fatalf("gate should reject non-matching tokens")
	}
}

func TestAssistedSearchGateInfo(t *testing.T) {
	var gate AssistedSearchGate
	if got := gate.Info(); got != "assisted search is disabled" {
		t.Fatalf("expected disabled info, got %q", got)
	}
	gate = gate.Enable("")
	if got := gate.Info(); got != "assisted search requires a consent token" {
		t.Fatalf("expected consent token hint, got %q", got)
	}
	gate = gate.Enable("consent-1")
	if got := gate.Info(); got != "assisted search gated" {
		t.Fatalf("expected gated info, got %q", got)
	}
}
