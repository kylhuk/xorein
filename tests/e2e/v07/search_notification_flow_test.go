package v07e2e

import (
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v07/notification"
	"github.com/aether/code_aether/pkg/v07/pushrelay"
	"github.com/aether/code_aether/pkg/v07/search"
)

func TestSearchNotificationFlow(t *testing.T) {
	cases := []struct {
		name                  string
		raw                   string
		filters               search.QueryFilters
		limit                 int
		offset                int
		wantNormalized        string
		wantPagination        search.Pagination
		scopeID               string
		actor                 string
		wantAuthorized        bool
		metadata              pushrelay.Metadata
		wantMetadataMinimal   bool
		tokenAge              time.Duration
		wantTokenStatus       pushrelay.TokenStatus
		wantTokenRotating     bool
		attempt               int
		wantBackoff           string
		dedupe                notification.DedupeWindow
		now                   time.Time
		wantSuppress          bool
		trigger               notification.TriggerType
		fallback              string
		wantAction            string
		wantSuppressionReason string
	}{
		{
			name: "desktop search path",
			raw:  "  archive file  ",
			filters: search.QueryFilters{
				FromUser: "user:alice",
				Range: [2]time.Time{
					time.Date(2025, 12, 30, 8, 0, 0, 0, time.UTC),
					time.Date(2025, 12, 30, 10, 0, 0, 0, time.UTC),
				},
				HasFile: true,
			},
			limit:               50,
			offset:              5,
			wantNormalized:      "archive file from:user:alice after:2025-12-30T08:00:00Z before:2025-12-30T10:00:00Z has:file",
			wantPagination:      search.Pagination{Limit: 50, Offset: 5},
			scopeID:             "scope-S7-feed",
			actor:               "user:alice",
			wantAuthorized:      true,
			metadata:            pushrelay.Metadata{"session": "abc", "channel": "feed"},
			wantMetadataMinimal: true,
			tokenAge:            15 * time.Minute,
			wantTokenStatus:     pushrelay.TokenStatusActive,
			wantTokenRotating:   false,
			attempt:             2,
			wantBackoff:         "backoff.short",
			dedupe: notification.DedupeWindow{
				LastFired: time.Date(2025, 12, 30, 11, 0, 0, 0, time.UTC),
				Interval:  30 * time.Minute,
			},
			now:                   time.Date(2025, 12, 30, 12, 0, 0, 0, time.UTC),
			wantSuppress:          false,
			trigger:               notification.TriggerDesktop,
			fallback:              "desktop:ack",
			wantAction:            "desktop:ack",
			wantSuppressionReason: "notification.desktop.ready",
		},
		{
			name:    "push retry dedupe suppression",
			raw:     "status",
			filters: search.QueryFilters{
				// intentionally minimal to exercise defaults
			},
			limit:          -1,
			offset:         -5,
			wantNormalized: "status",
			wantPagination: search.Pagination{Limit: 10, Offset: 0},
			scopeID:        "scope-other",
			actor:          "system:relay",
			wantAuthorized: false,
			metadata: pushrelay.Metadata{
				"k1": "v1",
				"k2": "v2",
				"k3": "v3",
				"k4": "v4",
			},
			wantMetadataMinimal: false,
			tokenAge:            90 * time.Minute,
			wantTokenStatus:     pushrelay.TokenStatusStale,
			wantTokenRotating:   false,
			attempt:             7,
			wantBackoff:         "backoff.long",
			dedupe: notification.DedupeWindow{
				LastFired: time.Date(2025, 12, 30, 12, 35, 55, 0, time.UTC),
				Interval:  15 * time.Second,
			},
			now:                   time.Date(2025, 12, 30, 12, 36, 0, 0, time.UTC),
			wantSuppress:          true,
			trigger:               notification.TriggerDesktop,
			fallback:              "",
			wantAction:            "desktop:show",
			wantSuppressionReason: "notification.desktop.suppressed",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			normalized, err := search.NormalizeQuery(tc.raw, tc.filters)
			if err != nil {
				t.Fatalf("normalize query failed: %v", err)
			}
			if normalized != tc.wantNormalized {
				t.Fatalf("normalized query mismatch: got %s", normalized)
			}

			pagination := search.NormalizePagination(tc.limit, tc.offset)
			if pagination != tc.wantPagination {
				t.Fatalf("pagination mismatch: got %+v", pagination)
			}

			if auth := search.ScopeAuthorized(tc.scopeID, tc.actor); auth != tc.wantAuthorized {
				t.Fatalf("scope authorization mismatch: want %v", tc.wantAuthorized)
			}

			if minimal := pushrelay.IsMetadataMinimal(tc.metadata); minimal != tc.wantMetadataMinimal {
				t.Fatalf("metadata minimalization mismatch: got %v", minimal)
			}

			lifecycle := pushrelay.EvaluateTokenLifecycle(tc.tokenAge)
			if lifecycle.Status != tc.wantTokenStatus {
				t.Fatalf("token status mismatch: got %s", lifecycle.Status)
			}
			if lifecycle.Rotating != tc.wantTokenRotating {
				t.Fatalf("token rotating mismatch: got %v", lifecycle.Rotating)
			}

			if backoff := pushrelay.BackoffClassification(tc.attempt); backoff != tc.wantBackoff {
				t.Fatalf("backoff mismatch: got %s", backoff)
			}

			suppressed := tc.dedupe.ShouldSuppress(tc.now)
			if suppressed != tc.wantSuppress {
				t.Fatalf("suppression mismatch: want %v", tc.wantSuppress)
			}

			action := notification.ResolveAction(tc.trigger, tc.fallback)
			if action != tc.wantAction {
				t.Fatalf("action mismatch: got %s", action)
			}

			reason := notification.SuppressionReason(tc.wantSuppress, tc.trigger)
			if reason != tc.wantSuppressionReason {
				t.Fatalf("suppression reason mismatch: got %s", reason)
			}
		})
	}
}
