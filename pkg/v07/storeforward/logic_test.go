package storeforward

import (
	"fmt"
	"math/rand"
	"testing"
	"testing/quick"
	"time"
)

func TestNormalizeTTLBounds(t *testing.T) {
	fallback := TTLRange{MinSeconds: 60, MaxSeconds: 300}
	if got := NormalizeTTL(0, fallback); got != 300*time.Second {
		t.Fatalf("expected fallback max, got %v", got)
	}
	if got := NormalizeTTL(5*time.Second, fallback); got != 60*time.Second {
		t.Fatalf("expected lower bound, got %v", got)
	}
	if got := NormalizeTTL(400*time.Second, fallback); got != 300*time.Second {
		t.Fatalf("expected upper bound, got %v", got)
	}
	if got := NormalizeTTL(120*time.Second, fallback); got != 120*time.Second {
		t.Fatalf("expected unchanged ttl, got %v", got)
	}
}

func TestRecordExpirationAndRemaining(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	record := Record{ID: "edge", CreatedAt: now, TTL: 5 * time.Minute, Metadata: map[string]string{}}
	if exp := record.Expiration(); !exp.Equal(now.Add(5 * time.Minute)) {
		t.Fatalf("unexpected expiration %v", exp)
	}
	if rem := record.Remaining(now.Add(2 * time.Minute)); rem != 3*time.Minute {
		t.Fatalf("unexpected remaining duration %v", rem)
	}
	if !record.IsExpired(now.Add(6 * time.Minute)) {
		t.Fatalf("expected record to be expired")
	}
	zeroTTL := Record{ID: "zero", CreatedAt: now, TTL: 0}
	if !zeroTTL.IsExpired(now) {
		t.Fatalf("zero ttl should expire immediately")
	}
}

func TestClassifyPurgeScenarios(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	record := Record{ID: "id", CreatedAt: now.Add(-10 * time.Minute), TTL: 5 * time.Minute, Metadata: map[string]string{}}
	cases := []struct {
		name            string
		record          Record
		storagePressure float64
		policyDays      int
		wantReason      PurgeReason
		expectDegraded  bool
	}{
		{
			name:            "expired",
			record:          record,
			storagePressure: 0,
			policyDays:      30,
			wantReason:      ReasonExpired,
		},
		{
			name:            "storage pressure",
			record:          Record{ID: "pressure", CreatedAt: now, TTL: 60 * time.Minute, Metadata: map[string]string{}},
			storagePressure: 0.85,
			policyDays:      0,
			wantReason:      ReasonStoragePressured,
		},
		{
			name:            "storage pressure degraded",
			record:          Record{ID: "pressure-degraded", CreatedAt: now, TTL: 60 * time.Minute, Metadata: map[string]string{}},
			storagePressure: 0.92,
			policyDays:      0,
			wantReason:      ReasonStoragePressured,
			expectDegraded:  true,
		},
		{
			name:            "policy enforced",
			record:          Record{ID: "policy", CreatedAt: now.Add(-40 * 24 * time.Hour), TTL: 90 * 24 * time.Hour, Metadata: map[string]string{}},
			storagePressure: 0,
			policyDays:      30,
			wantReason:      ReasonPolicyEnforced,
		},
		{
			name:            "degraded signal indicator",
			record:          Record{ID: "degraded", CreatedAt: now, TTL: 60 * time.Minute, Metadata: map[string]string{"replication": "degraded"}},
			storagePressure: 0.2,
			policyDays:      0,
			wantReason:      ReasonDegradedSignal,
			expectDegraded:  true,
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			class := ClassifyPurge(tc.record, now, tc.policyDays, tc.storagePressure)
			if class.Reason != tc.wantReason {
				t.Fatalf("unexpected reason %s", class.Reason)
			}
			if class.RecordID != tc.record.ID {
				t.Fatalf("unexpected record id %s", class.RecordID)
			}
			if class.IsDegraded != tc.expectDegraded {
				t.Fatalf("unexpected degraded flag %v", class.IsDegraded)
			}
		})
	}
}

func TestClassifyPurgeDeterministic(t *testing.T) {
	source := rand.NewSource(99)
	rng := rand.New(source)
	now := time.Now().UTC()
	record := Record{ID: "deterministic", CreatedAt: now, TTL: time.Hour, Metadata: map[string]string{}}
	base := ClassifyPurge(record, now, 0, 0)
	for i := 0; i < 10; i++ {
		jitter := time.Duration(rng.Intn(1000)) * time.Millisecond
		newRecord := record
		newRecord.CreatedAt = now.Add(-jitter)
		repeated := ClassifyPurge(newRecord, now, 0, 0)
		if repeated.Reason != base.Reason {
			t.Fatalf("inconsistent reason %s vs %s", repeated.Reason, base.Reason)
		}
		expectedNext := newRecord.Expiration()
		if !repeated.NextCheck.Equal(expectedNext) {
			t.Fatalf("unexpected next check %v vs expected %v", repeated.NextCheck, expectedNext)
		}
	}
}

func TestClassifyPurgeRetainVsExpireProperty(t *testing.T) {
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	prop := func(ageMinutes uint8) bool {
		age := int(ageMinutes % 120)
		record := Record{
			ID:        fmt.Sprintf("prop-%d", age),
			CreatedAt: now.Add(-time.Duration(age) * time.Minute),
			TTL:       45 * time.Minute,
			Metadata:  map[string]string{},
		}
		classification := ClassifyPurge(record, now, 0, 0)
		if age >= 45 {
			return classification.Reason == ReasonExpired
		}
		return classification.Reason == ReasonRetained
	}

	if err := quick.Check(prop, nil); err != nil {
		t.Fatalf("retain-vs-expire property failed: %v", err)
	}
}

func TestBuildReplicationPlan(t *testing.T) {
	peers := make([]PeerInfo, 25)
	for i := range peers {
		peers[i] = PeerInfo{ID: fmt.Sprintf("peer-%02d", i), Score: i % 5}
	}
	record := Record{ID: "rp"}
	plan := BuildReplicationPlan(record, peers)
	if plan.RecordID != record.ID {
		t.Fatalf("unexpected record id")
	}
	if len(plan.Targets) != ReplicationTarget {
		t.Fatalf("unexpected target count %d", len(plan.Targets))
	}
	if plan.Reason != "replication.target.met" {
		t.Fatalf("unexpected reason %s", plan.Reason)
	}
	if plan.NeedsRepair() {
		t.Fatalf("did not expect repair")
	}
	plan = BuildReplicationPlan(record, peers[:10])
	if plan.Reason != fmt.Sprintf("replication.degraded.target=%d", ReplicationTarget) {
		t.Fatalf("unexpected degraded reason %s", plan.Reason)
	}
	if !plan.NeedsRepair() {
		t.Fatalf("expected repair")
	}
}

func TestBuildReplicationPlanFanoutOrdering(t *testing.T) {
	peers := make([]PeerInfo, ReplicationTarget+5)
	for i := range peers {
		peers[i] = PeerInfo{ID: fmt.Sprintf("fan-%02d", i), Score: ReplicationTarget - i}
	}
	record := Record{ID: "fanout"}
	plan := BuildReplicationPlan(record, peers)
	if len(plan.Targets) != ReplicationTarget {
		t.Fatalf("expected %d targets, got %d", ReplicationTarget, len(plan.Targets))
	}
	if plan.Targets[0].Score < plan.Targets[len(plan.Targets)-1].Score {
		t.Fatalf("expected descending score order")
	}
	if plan.Reason != "replication.target.met" {
		t.Fatalf("unexpected reason %s", plan.Reason)
	}
}
