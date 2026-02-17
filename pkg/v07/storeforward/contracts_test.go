package storeforward

import (
	"testing"
	"time"
)

func TestTTLPolicyValid(t *testing.T) {
	policy := TTLPolicy{Min: time.Minute, Max: 10 * time.Minute}
	if !policy.IsValid(5 * time.Minute) {
		t.Fatalf("expected TTL within range to be valid")
	}
	if policy.IsValid(time.Second) {
		t.Fatalf("expected TTL below min to be invalid")
	}
}

func TestAssessReplication(t *testing.T) {
	healthy := AssessReplication(3, 3)
	if healthy.Status != ReplicationStatusHealthy {
		t.Fatalf("expected healthy status, got %s", healthy.Status)
	}
	degraded := AssessReplication(1, 3)
	if degraded.Status != ReplicationStatusDegraded {
		t.Fatalf("expected degraded status, got %s", degraded.Status)
	}
	critical := AssessReplication(0, 5)
	if critical.Status != ReplicationStatusCritical {
		t.Fatalf("expected critical status, got %s", critical.Status)
	}
	if !NeedsRetry(degraded) || !NeedsRetry(critical) {
		t.Fatalf("expected degraded and critical to need retry")
	}
}

func TestBoundedRetryDelay(t *testing.T) {
	if BoundedRetryDelay(0) != 0 {
		t.Fatalf("expected zero delay for attempt 0")
	}
	if d := BoundedRetryDelay(100); d != 5*time.Second {
		t.Fatalf("expected capped delay 5s, got %v", d)
	}
}
