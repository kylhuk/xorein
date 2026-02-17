package search

import (
	"reflect"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v06/conformance"
)

func TestSearchContractHelpers(t *testing.T) {
	tests := []struct {
		name       string
		contract   SearchContract
		wantAnchor string
		wantReason string
	}{
		{
			name:       "partial",
			contract:   NewSearchContract("S6-03", "T6-10", "search-doc", conformance.GateV6G1, SearchReasonPartial),
			wantAnchor: "search-doc#T6-10",
			wantReason: string(SearchReasonPartial),
		},
		{
			name:       "preview",
			contract:   NewSearchContract("S6-05", "T6-11", "preview-doc", conformance.GateV6G3, SearchReasonPreview),
			wantAnchor: "preview-doc#T6-11",
			wantReason: string(SearchReasonPreview),
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

func TestSearchReasonClassesDeterministic(t *testing.T) {
	want := []SearchReasonClass{SearchReasonPartial, SearchReasonExplore, SearchReasonPreview}
	got := SearchReasonClasses()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("reason classes changed: want %v got %v", want, got)
	}
}

func TestClassifyPartialFailure(t *testing.T) {
	tests := []struct {
		name       string
		successful int
		total      int
		wantStatus PartialFailureStatus
		wantReason string
		needRetry  bool
	}{
		{name: "success", successful: 5, total: 5, wantStatus: PartialFailureSuccess, wantReason: string(PartialFailureSuccess), needRetry: false},
		{name: "partial", successful: 2, total: 5, wantStatus: PartialFailureFailure, wantReason: string(PartialFailureFailure), needRetry: true},
		{name: "fallback", successful: 0, total: 3, wantStatus: PartialFailureFallback, wantReason: string(PartialFailureFallback), needRetry: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyPartialFailure(tt.successful, tt.total)
			if got.Status != tt.wantStatus || got.ReasonLabel != tt.wantReason || got.NeedsFallback != tt.needRetry {
				t.Fatalf("partial failure mismatch: want %s/%s/%t got %s/%s/%t", tt.wantStatus, tt.wantReason, tt.needRetry, got.Status, got.ReasonLabel, got.NeedsFallback)
			}
		})
	}
}

func TestOrderExploreFeed(t *testing.T) {
	now := time.Now()
	items := []FeedItem{
		{ID: "first", Freshness: now.Add(-5 * time.Minute), Priority: 1},
		{ID: "second", Freshness: now.Add(-1 * time.Minute), Priority: 2},
		{ID: "third", Freshness: now.Add(-1 * time.Minute), Priority: 0},
	}
	got := OrderExploreFeed(items, now.Add(-2*time.Minute))
	if got[0].ID != "first" || got[1].ID != "third" || got[2].ID != "second" {
		t.Fatalf("ordering mismatch: got %v", []string{got[0].ID, got[1].ID, got[2].ID})
	}
	if got[1].Priority != 1 || got[2].Priority != 3 {
		t.Fatalf("priority bump mismatch")
	}
}

func TestDecidePreviewMismatch(t *testing.T) {
	tests := []struct {
		name       string
		expected   string
		actual     string
		wantMatch  bool
		wantReason string
	}{
		{name: "aligned", expected: "foo", actual: "foo", wantMatch: true, wantReason: string(SearchReasonPreview) + ".success"},
		{name: "mismatch", expected: "foo", actual: "bar", wantMatch: false, wantReason: string(SearchReasonPreview) + ".mismatch"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := DecidePreviewMismatch(tt.expected, tt.actual)
			if got.Matched != tt.wantMatch || got.Reason != tt.wantReason {
				t.Fatalf("preview mismatch: want %t/%s got %t/%s", tt.wantMatch, tt.wantReason, got.Matched, got.Reason)
			}
		})
	}
}
