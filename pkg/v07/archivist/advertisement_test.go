package archivist

import (
	"testing"
	"time"
)

func TestSelectArchivistsDeterministic(t *testing.T) {
	candidates := []CapabilityAnnouncement{
		{PeerID: "peer-b", Score: 5},
		{PeerID: "peer-a", Score: 5},
		{PeerID: "peer-c", Score: 4},
	}
	result := SelectArchivists(candidates, 2)
	if len(result) != 2 {
		t.Fatalf("expected two archivists, got %d", len(result))
	}
	if result[0].PeerID != "peer-a" {
		t.Fatalf("expected peer-a first, got %s", result[0].PeerID)
	}
	if result[1].PeerID != "peer-b" {
		t.Fatalf("expected peer-b second, got %s", result[1].PeerID)
	}
}

func TestCoverageDropSignalsFallback(t *testing.T) {
	signal := CoverageDrop("arch-1", 0.3)
	if signal.Message != "archivist.coverage.low=0.30" {
		t.Fatalf("unexpected message %s", signal.Message)
	}
	signal = CoverageDrop("arch-1", 0.6)
	if signal.Message != "archivist.coverage=0.60" {
		t.Fatalf("unexpected message %s", signal.Message)
	}
}

func TestGracefulWithdrawalState(t *testing.T) {
	withdrawn := GracefulWithdrawal("peer-x")
	if withdrawn.State != StateWithdrawn {
		t.Fatalf("expected withdrawn state, got %s", withdrawn.State)
	}
	if withdrawn.Score != 0 {
		t.Fatalf("expected zero score, got %d", withdrawn.Score)
	}
	if withdrawn.Expires.Before(time.Now().UTC()) {
		t.Fatalf("expected future expiry")
	}
}
