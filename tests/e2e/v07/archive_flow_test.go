package v07e2e

import (
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v07/archivist"
	"github.com/aether/code_aether/pkg/v07/history"
	"github.com/aether/code_aether/pkg/v07/retention"
	"github.com/aether/code_aether/pkg/v07/storeforward"
)

func TestArchiveFlow(t *testing.T) {
	policy := storeforward.TTLPolicy{Min: time.Minute, Max: 30 * time.Minute}

	cases := []struct {
		name             string
		ttl              time.Duration
		expectTTLExpired bool
		retentionCurrent retention.RetentionPolicy
		retentionTarget  retention.RetentionPolicy
		wantAction       retention.TransitionAction
		wantPurgeReason  string
		replicaCount     int
		expectedReplicas int
		wantReplication  storeforward.ReplicationStatus
		expectRetry      bool
		proofExpected    string
		proofActual      string
		expectProofMatch bool
		initialState     archivist.EnrollmentState
		nextState        archivist.EnrollmentState
		archivistReason  string
		expectRecovery   bool
	}{
		{
			name:             "healthy archive path",
			ttl:              5 * time.Minute,
			expectTTLExpired: false,
			retentionCurrent: retention.RetentionPolicy{Tier: retention.TierEdge, Days: 14},
			retentionTarget:  retention.RetentionPolicy{Tier: retention.TierCore, Days: 30},
			wantAction:       retention.TransitionArchive,
			replicaCount:     4,
			expectedReplicas: 4,
			wantReplication:  storeforward.ReplicationStatusHealthy,
			expectRetry:      false,
			proofExpected:    "root-match",
			proofActual:      "root-match",
			expectProofMatch: true,
			initialState:     archivist.StateEnrolling,
			nextState:        archivist.StateActive,
			archivistReason:  "enrollment complete",
			expectRecovery:   false,
		},
		{
			name:             "degraded purge recovery",
			ttl:              -1,
			expectTTLExpired: true,
			retentionCurrent: retention.RetentionPolicy{Tier: retention.TierCore, Days: 30},
			retentionTarget:  retention.RetentionPolicy{Tier: retention.TierBurst, Days: 7},
			wantAction:       retention.TransitionPurge,
			wantPurgeReason:  "burst purge",
			replicaCount:     1,
			expectedReplicas: 3,
			wantReplication:  storeforward.ReplicationStatusDegraded,
			expectRetry:      true,
			proofExpected:    "root-A",
			proofActual:      "root-B",
			expectProofMatch: false,
			initialState:     archivist.StateActive,
			nextState:        archivist.StateSuspended,
			archivistReason:  "replication degraded",
			expectRecovery:   true,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			valid := policy.IsValid(tt.ttl)
			if tt.expectTTLExpired && valid {
				t.Fatalf("expected TTL expiry for %s", tt.name)
			}
			if !tt.expectTTLExpired && !valid {
				t.Fatalf("unexpected TTL expiry for %s", tt.name)
			}

			action := retention.DetermineTransition(tt.retentionCurrent, tt.retentionTarget)
			if action != tt.wantAction {
				t.Fatalf("expected %s, got %s", tt.wantAction, action)
			}
			if tt.wantPurgeReason != "" {
				plan := retention.BuildPurgePlan(tt.retentionTarget)
				if plan.Reason != tt.wantPurgeReason {
					t.Fatalf("expected purge reason %s, got %s", tt.wantPurgeReason, plan.Reason)
				}
			}

			assessment := storeforward.AssessReplication(tt.replicaCount, tt.expectedReplicas)
			if assessment.Status != tt.wantReplication {
				t.Fatalf("replication status mismatch: got %s", assessment.Status)
			}
			if needRetry := storeforward.NeedsRetry(assessment); needRetry != tt.expectRetry {
				t.Fatalf("retry expectation mismatch: want %v, got %v", tt.expectRetry, needRetry)
			}

			proof := history.ClassifyProof(tt.proofExpected, tt.proofActual)
			if proof.Matched != tt.expectProofMatch {
				t.Fatalf("proof match mismatch: expected %v", tt.expectProofMatch)
			}
			wantProofReason := "proof.match"
			if !tt.expectProofMatch {
				wantProofReason = "proof.mismatch"
			}
			if proof.Reason != wantProofReason {
				t.Fatalf("unexpected proof reason: got %s", proof.Reason)
			}

			capsule := history.NewCapsuleMetadata("v07-archive", 13, history.ModeEpochHistory)
			if capsule.Root != history.CanonicalRoot("v07-archive", 13, history.ModeEpochHistory) {
				t.Fatalf("canonical root mismatch: got %s", capsule.Root)
			}
			sync := history.ResumeSync(history.SyncState{Cursor: "cursor-" + tt.name, Epoch: capsule.Epoch})
			if !sync.Ready {
				t.Fatalf("expected sync ready after resume")
			}

			if !archivist.CanTransition(tt.initialState, tt.nextState) {
				t.Fatalf("invalid transition %s -> %s", tt.initialState, tt.nextState)
			}
			rec := archivist.RecordTransition(tt.initialState, tt.nextState, tt.archivistReason)
			if rec.Reason != tt.archivistReason {
				t.Fatalf("unexpected archivist reason: got %s", rec.Reason)
			}

			if tt.expectRecovery {
				if !archivist.CanTransition(rec.To, archivist.StateActive) {
					t.Fatalf("recovery transition not allowed from %s", rec.To)
				}
				recovery := archivist.RecordTransition(rec.To, archivist.StateActive, "recovery")
				if recovery.Reason != "recovery" {
					t.Fatalf("unexpected recovery reason: got %s", recovery.Reason)
				}
				if recovery.To != archivist.StateActive {
					t.Fatalf("expected recovery to active, got %s", recovery.To)
				}
			}
		})
	}
}
