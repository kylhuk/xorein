package v08e2e

import (
	"testing"

	"github.com/aether/code_aether/pkg/v08/linkpreview"
	"github.com/aether/code_aether/pkg/v08/threads"
)

func TestThreadsPreviewFlow(t *testing.T) {
	cases := []struct {
		name          string
		trace         threads.ThreadTrace
		wantLifecycle threads.Lifecycle
		rawURL        string
		wantEligible  bool
		wantErr       bool
	}{
		{
			name:          "valid thread",
			trace:         threads.ThreadTrace{ID: "flow", CreatedDepth: 1, ReplyDepth: 2},
			wantLifecycle: threads.LifecycleActive,
			rawURL:        "https://example.com/tmp",
			wantEligible:  true,
		},
		{
			name:          "invalid reply",
			trace:         threads.ThreadTrace{ID: "flow", CreatedDepth: 3, ReplyDepth: 6},
			wantLifecycle: threads.LifecycleArchived,
			rawURL:        "http://example.com",
			wantEligible:  false,
			wantErr:       true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := threads.ValidateReplyLineage(tc.trace)
			if (err != nil) != tc.wantErr {
				t.Fatalf("unexpected error state: %v", err)
			}
			if got := threads.ClassifyLifecycle(tc.trace.ReplyDepth); got != tc.wantLifecycle {
				t.Fatalf("lifecycle mismatch: %s", got)
			}
			normalized, err := linkpreview.NormalizeURL(tc.rawURL)
			if err != nil {
				t.Fatalf("normalize failed: %v", err)
			}
			if linkpreview.PreviewEligibility(normalized) != tc.wantEligible {
				t.Fatalf("eligibility mismatch for %s", normalized)
			}
		})
	}
}
