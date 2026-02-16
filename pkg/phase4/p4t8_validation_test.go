package phase4

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestP4T8RepeatedCleanStatePass(t *testing.T) {
	for run := 1; run <= 2; run++ {
		t.Run(fmt.Sprintf("clean-run-%d", run), func(t *testing.T) {
			lan := newDeterministicLAN(t, []string{"node-a", "node-b"})
			lan.simulateDiscovery()
			lan.assertPeerLists()

			connectNodeFromDiscoveredPeer(t, "p4t8-clean-state", "node-a", lan.nodes[0].discovery.KnownPeers())
			connectNodeFromDiscoveredPeer(t, "p4t8-clean-state", "node-b", lan.nodes[1].discovery.KnownPeers())
		})
	}
}

func connectNodeFromDiscoveredPeer(t *testing.T, namespace, node string, discovered []string) {
	t.Helper()
	if len(discovered) != 1 {
		t.Fatalf("%s expected exactly one discovered peer, got %v", node, discovered)
	}
	reachablePeer := discovered[0]
	cfg := DHTConfig{
		Namespace:      namespace,
		BootstrapPeers: discovered,
	}
	probe := func(_ context.Context, peer string) bool {
		return peer == reachablePeer
	}
	svc, err := NewDHTService(cfg, probe)
	if err != nil {
		t.Fatalf("%s NewDHTService: %v", node, err)
	}
	statuses := svc.Bootstrap(context.Background())
	if len(statuses) != 1 {
		t.Fatalf("%s unexpected status count = %d", node, len(statuses))
	}
	if !statuses[0].Success || statuses[0].Peer != reachablePeer {
		t.Fatalf("%s expected bootstrap success on %s, got %#v", node, reachablePeer, statuses[0])
	}
	if statuses[0].Reason != "joined namespace" {
		t.Fatalf("%s unexpected reason = %q", node, statuses[0].Reason)
	}
}

func TestP4T8RestartReconnectUsesPeerCache(t *testing.T) {
	dir := t.TempDir()
	cachePath := filepath.Join(dir, "peer-cache.json")
	firstPeer := "cached-peer"

	cfg := DHTConfig{
		Namespace:      "p4t8-reconnect",
		BootstrapPeers: []string{firstPeer},
		PeerCachePath:  cachePath,
	}

	var firstOrder []string
	probeFirst := func(_ context.Context, peer string) bool {
		firstOrder = append(firstOrder, peer)
		return peer == firstPeer
	}
	svc, err := NewDHTService(cfg, probeFirst)
	if err != nil {
		t.Fatalf("NewDHTService: %v", err)
	}
	if len(svc.Bootstrap(context.Background())) != 1 {
		t.Fatalf("bootstrap failed on first run")
	}
	if len(firstOrder) != 1 || firstOrder[0] != firstPeer {
		t.Fatalf("expected single bootstrap attempt on %q, got %v", firstPeer, firstOrder)
	}

	secondCfg := DHTConfig{
		Namespace:      "p4t8-reconnect",
		BootstrapPeers: []string{"bootstrap-fallback"},
		PeerCachePath:  cachePath,
	}
	var secondOrder []string
	probeSecond := func(_ context.Context, peer string) bool {
		secondOrder = append(secondOrder, peer)
		return peer == firstPeer
	}
	svc2, err := NewDHTService(secondCfg, probeSecond)
	if err != nil {
		t.Fatalf("NewDHTService second: %v", err)
	}
	statuses := svc2.Bootstrap(context.Background())
	if len(statuses) == 0 || !statuses[0].Success {
		t.Fatalf("expected cached peer success, got %#v", statuses)
	}
	if len(secondOrder) == 0 {
		t.Fatalf("no peers probed on restart")
	}
	if secondOrder[0] != firstPeer {
		t.Fatalf("cached peer was not probed first: %v", secondOrder)
	}
}

func TestP4T8TraversalFailureObservations(t *testing.T) {
	hooks := TraversalHooks{
		StageActions: map[TraversalStage]StageAction{
			TraversalStageDirect:    failStageAction(ReasonDirectUnavailable),
			TraversalStageAutoNAT:   failStageAction(ReasonAutoNATFailure),
			TraversalStageHolePunch: failStageAction(ReasonHolePunchFailure),
		},
		RelayAction: func(_ context.Context) RelayReservationOutcome {
			return RelayReservationOutcome{
				State: RelayReservationStateFailed,
				Events: []RelayReservationEvent{{
					State:   RelayReservationStateFailed,
					Reason:  ReasonRelayReservationFailed,
					Message: "relay hop blocked",
				}},
			}
		},
	}
	report := NewTraversalRunner(FallbackTimeoutPolicy{}, hooks).Run(context.Background())
	if report.Stage != TraversalStageRelay {
		t.Fatalf("expected final stage relay, got %s", report.Stage)
	}
	if report.Reason != ReasonRelayReservationFailed {
		t.Fatalf("expected relay failure reason, got %s", report.Reason)
	}
	if len(report.Events) != len(defaultTraversalOrder) {
		t.Fatalf("unexpected event count %d", len(report.Events))
	}

	expected := []struct {
		stage  TraversalStage
		reason ConnectivityReasonCode
	}{
		{TraversalStageDirect, ReasonDirectUnavailable},
		{TraversalStageAutoNAT, ReasonAutoNATFailure},
		{TraversalStageHolePunch, ReasonHolePunchFailure},
		{TraversalStageRelay, ReasonRelayReservationFailed},
	}

	for i, want := range expected {
		got := report.Events[i]
		if got.Stage != want.stage {
			t.Fatalf("event[%d] stage = %s, want %s", i, got.Stage, want.stage)
		}
		if got.Reason != want.reason {
			t.Fatalf("event[%d] reason = %s, want %s", i, got.Reason, want.reason)
		}
		if !strings.Contains(strings.ToLower(got.Message), strings.ToLower(string(want.stage))) {
			t.Fatalf("event[%d] message %q must mention %s", i, got.Message, want.stage)
		}
	}
}

func failStageAction(code ConnectivityReasonCode) StageAction {
	return func(_ context.Context) StageResult {
		return StageResult{
			Success: false,
			Reason:  code,
			Message: fmt.Sprintf("simulated packet loss %s", code),
		}
	}
}
