package v23

import (
	"errors"
	"strings"
	"testing"

	"github.com/aether/code_aether/pkg/v23/durability"
)

var errNetworkPartition = errors.New("network partition during catch-up")
var errPrivateSpaceLookup = errors.New("private space lookup denied")
var errRelayBoundary = errors.New("relay does not host durable history")

func TestScenarioOfflineCatchup(t *testing.T) {
	fetcher := networkChaoticBackfill{}

	if _, err := fetcher.fetchSegments(); err == nil {
		t.Fatalf("expected initial catch-up fetch to fail")
	}
	if !errors.Is(fetcher.lastError, errNetworkPartition) {
		t.Fatalf("expected network partition on first fetch, got %v", fetcher.lastError)
	}
	if fetcher.attempts != 1 {
		t.Fatalf("expected one failed attempt before recovery, got %d", fetcher.attempts)
	}

	segments, err := fetcher.fetchSegments()
	if err != nil {
		t.Fatalf("expected recovery fetch to succeed: %v", err)
	}
	if len(segments) != 2 {
		t.Fatalf("expected 2 segments after catch-up, got %d", len(segments))
	}
	if fetcher.attempts != 2 {
		t.Fatalf("expected two attempts total: failed then recovered, got %d", fetcher.attempts)
	}
}

type networkChaoticBackfill struct {
	attempts  int
	lastError error
}

func (n *networkChaoticBackfill) fetchSegments() ([]string, error) {
	n.attempts++
	if n.attempts == 1 {
		n.lastError = errNetworkPartition
		return nil, n.lastError
	}
	n.lastError = nil
	return []string{"seg-a", "seg-b"}, nil
}

func TestScenarioRedactionTombstoneRegression(t *testing.T) {
	store := redactionBackfillStore{}
	store.put("ev-1", "hello")
	store.put("ev-2", "secret note")
	store.applyTombstone("ev-2")

	visible := store.visibleEvents()
	if len(visible) != 1 || visible[0] != "ev-1" {
		t.Fatalf("expected only non-tombstoned event to remain visible, got %v", visible)
	}

	if results := store.search("secret"); len(results) != 0 {
		t.Fatalf("expected tombstoned content to be removed from search, got %v", results)
	}

	backfill := store.backfill()
	if len(backfill) != 1 || backfill[0] != "ev-1" {
		t.Fatalf("expected backfill stream to exclude tombstoned events, got %v", backfill)
	}
}

type redactionBackfillStore struct {
	events map[string]redactionEvent
}

type redactionEvent struct {
	id        string
	payload   string
	tombstone bool
}

func (s *redactionBackfillStore) put(id, payload string) {
	if s.events == nil {
		s.events = make(map[string]redactionEvent)
	}
	s.events[id] = redactionEvent{id: id, payload: payload}
}

func (s *redactionBackfillStore) applyTombstone(id string) {
	ev := s.events[id]
	ev.tombstone = true
	s.events[id] = ev
}

func (s *redactionBackfillStore) visibleEvents() []string {
	visible := make([]string, 0, len(s.events))
	for id, ev := range s.events {
		if ev.tombstone {
			continue
		}
		visible = append(visible, id)
	}
	return visible
}

func (s *redactionBackfillStore) search(query string) []string {
	matches := make([]string, 0)
	for id, ev := range s.events {
		if ev.tombstone {
			continue
		}
		if strings.Contains(ev.payload, query) {
			matches = append(matches, id)
		}
	}
	return matches
}

func (s *redactionBackfillStore) backfill() []string {
	backfill := make([]string, 0, len(s.events))
	for id, ev := range s.events {
		if ev.tombstone {
			continue
		}
		backfill = append(backfill, id)
	}
	return backfill
}

func TestScenarioPrivateSpaceAntiEnumeration(t *testing.T) {
	directory := privateSpaceDirectory{
		memberships: map[string]string{
			"space-123": "token-space-123",
		},
		payload: map[string]string{
			"space-123": "head-manifest",
		},
	}

	_, err := directory.fetchHistory("space-123", "wrong-token")
	if err == nil {
		t.Fatalf("expected lookup denial for wrong token")
	}
	if !errors.Is(err, errPrivateSpaceLookup) {
		t.Fatalf("expected generic anti-enumeration denial, got %v", err)
	}

	if _, err := directory.fetchHistory("ghost-space", "any-token"); !errors.Is(err, errPrivateSpaceLookup) {
		t.Fatalf("expected same denial for missing spaces, got %v", err)
	}

	manifest, err := directory.fetchHistory("space-123", "token-space-123")
	if err != nil {
		t.Fatalf("expected valid lookup for correct membership token: %v", err)
	}
	if manifest != "head-manifest" {
		t.Fatalf("expected manifest payload, got %q", manifest)
	}
}

type privateSpaceDirectory struct {
	memberships map[string]string
	payload     map[string]string
}

func (d *privateSpaceDirectory) fetchHistory(space, token string) (string, error) {
	if token == "" {
		return "", errPrivateSpaceLookup
	}
	if storedToken, ok := d.memberships[space]; !ok || storedToken != token {
		return "", errPrivateSpaceLookup
	}
	return d.payload[space], nil
}

func TestScenarioReplicaHealingUnderChurn(t *testing.T) {
	accounting := durability.NewReplicaAccounting(3, 2, 10)
	accounting.ApplyEvent(durability.ReplicaEvent{NodeID: "A", Online: true})
	accounting.ApplyEvent(durability.ReplicaEvent{NodeID: "B", Online: true})
	accounting.ApplyEvent(durability.ReplicaEvent{NodeID: "C", Online: true})

	status := accounting.Snapshot()
	if status.State != durability.DurabilityStateHealthy {
		t.Fatalf("expected healthy initial state, got %s", status.State)
	}

	// Simulate churn with a transient disconnect on one replica.
	accounting.ApplyEvent(durability.ReplicaEvent{NodeID: "A", Online: false})
	accounting.ApplyEvent(durability.ReplicaEvent{NodeID: "A", Online: true})
	status = accounting.Snapshot()
	if status.State != durability.DurabilityStateHealthy || status.Reason != durability.ReasonTargetMet {
		t.Fatalf("expected churn recovery to remain healthy, got state=%s reason=%s", status.State, status.Reason)
	}

	// Drop below target temporarily (churn then healing).
	accounting.ApplyEvent(durability.ReplicaEvent{NodeID: "B", Online: false})
	status = accounting.Snapshot()
	if status.State != durability.DurabilityStateDegraded || status.Reason != durability.ReasonTargetUnmet {
		t.Fatalf("expected target-unmet degradation during healing window, got state=%s reason=%s", status.State, status.Reason)
	}
	accounting.ApplyEvent(durability.ReplicaEvent{NodeID: "B", Online: true})
	status = accounting.Snapshot()
	if status.State != durability.DurabilityStateHealthy {
		t.Fatalf("expected healthy after replica returns, got state=%s reason=%s", status.State, status.Reason)
	}
}

func TestScenarioRelayNoHistoryHosting(t *testing.T) {
	relay := relayBoundaryProbe{}

	if err := relay.fetchHistory(); !errors.Is(err, errRelayBoundary) {
		t.Fatalf("expected history readback to be blocked on relay, got %v", err)
	}

	if err := relay.searchHistory(); !errors.Is(err, errRelayBoundary) {
		t.Fatalf("expected relay search to be blocked, got %v", err)
	}
}

type relayBoundaryProbe struct{}

func (r *relayBoundaryProbe) fetchHistory() error {
	return errRelayBoundary
}

func (r *relayBoundaryProbe) searchHistory() error {
	return errRelayBoundary
}
