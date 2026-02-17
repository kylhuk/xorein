package reporting

import (
	"reflect"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

func TestReportingContractHelpers(t *testing.T) {
	tests := []struct {
		name       string
		contract   ReportingContract
		wantAnchor string
		wantReason string
	}{
		{
			name:       "route",
			contract:   NewReportingContract("S6-09", "T6-40", "route-doc", conformance.GateV6G4, ReportingReasonRoute),
			wantAnchor: "route-doc#T6-40",
			wantReason: string(ReportingReasonRoute),
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

func TestReportingReasonClassesDeterministic(t *testing.T) {
	want := []ReportingReasonClass{ReportingReasonRoute}
	got := ReportingReasonClasses()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reason classes changed: want %v got %v", want, got)
	}
}

func TestComputeIdempotencyKeyStable(t *testing.T) {
	const want = "fb6b7f2400ea50fa618b3ad07475d741adf7f99c83c18a44cc9e65835b5f25c7"
	if got := computeIdempotencyKey("report-123"); got != want {
		t.Fatalf("idempotency key mismatch: want %s got %s", want, got)
	}
}

func TestEvaluateReportRouting(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name         string
		reportID     string
		previous     []string
		acked        bool
		orderingHint int
		lastAck      time.Time
		wantAccept   bool
		wantRetry    bool
		wantReason   string
	}{
		{name: "recent ack", reportID: "R1", acked: true, orderingHint: 0, lastAck: now.Add(-30 * time.Second), wantAccept: true, wantRetry: false, wantReason: "reporting.route.accept"},
		{name: "duplicate", reportID: "dup", previous: []string{"dup", "other"}, acked: false, lastAck: now.Add(-5 * time.Minute), wantAccept: false, wantRetry: true, wantReason: "reporting.route.failure"},
		{name: "ordering retry", reportID: "new", previous: []string{"a", "b", "c"}, orderingHint: 1, lastAck: now.Add(-5 * time.Minute), wantAccept: true, wantRetry: true, wantReason: "reporting.route.retry"},
		{name: "clean new", reportID: "fresh", previous: []string{"x"}, orderingHint: 5, lastAck: now.Add(-5 * time.Minute), wantAccept: true, wantRetry: false, wantReason: "reporting.route.accept"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := EvaluateReportRouting(tt.reportID, tt.previous, tt.acked, tt.orderingHint, tt.lastAck)
			if got.Accept != tt.wantAccept || got.Retry != tt.wantRetry || got.Reason != tt.wantReason {
				t.Fatalf("routing mismatch: want %t/%t/%s got %t/%t/%s", tt.wantAccept, tt.wantRetry, tt.wantReason, got.Accept, got.Retry, got.Reason)
			}
		})
	}
}
