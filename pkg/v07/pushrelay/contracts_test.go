package pushrelay

import (
	"testing"
	"time"
)

func TestIsMetadataMinimal(t *testing.T) {
	meta := Metadata{"a": "1", "b": "2"}
	if !IsMetadataMinimal(meta) {
		t.Fatalf("expected metadata size small enough")
	}
	meta["c"] = "3"
	meta["d"] = "4"
	if IsMetadataMinimal(meta) {
		t.Fatalf("expected metadata to exceed minimal threshold")
	}
}

func TestEvaluateTokenLifecycle(t *testing.T) {
	life := EvaluateTokenLifecycle(10 * time.Minute)
	if life.Status != TokenStatusActive {
		t.Fatalf("expected active status, got %s", life.Status)
	}
	rotating := EvaluateTokenLifecycle(40 * time.Minute)
	if rotating.Status != TokenStatusRotating || !rotating.Rotating {
		t.Fatalf("expected rotating status")
	}
	stale := EvaluateTokenLifecycle(90 * time.Minute)
	if stale.Status != TokenStatusStale || stale.Rotating {
		t.Fatalf("expected stale status without rotating")
	}
}

func TestBackoffClassification(t *testing.T) {
	if BackoffClassification(0) != "backoff.immediate" {
		t.Fatalf("unexpected classification for attempt 0")
	}
	if BackoffClassification(5) != "backoff.medium" {
		t.Fatalf("unexpected classification for attempt 5")
	}
}
