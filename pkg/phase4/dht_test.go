package phase4

import (
	"context"
	"testing"
)

func TestDHTServiceBootstrap(t *testing.T) {
	t.Run("joins normalized peer", func(t *testing.T) {
		cfg := DHTConfig{
			Namespace:      "phase4-test",
			BootstrapPeers: []string{"peer-b", "peer-a"},
		}
		svc, err := NewDHTService(cfg, func(_ context.Context, peer string) bool {
			return peer == "peer-b"
		})
		if err != nil {
			t.Fatalf("expected service, got error: %v", err)
		}

		statuses := svc.Bootstrap(context.Background())
		if len(statuses) != 2 {
			t.Fatalf("unexpected status count: %d", len(statuses))
		}
		if statuses[0].Success {
			t.Fatalf("expected first normalized peer to fail")
		}
		if !statuses[1].Success || statuses[1].Peer != "peer-b" {
			t.Fatalf("expected second peer to succeed, got %+v", statuses[1])
		}

		if peers := svc.KnownPeers(); !containsPeer(peers, "peer-b") {
			t.Fatalf("expected known peers to include joined node")
		}
		if got := svc.LastBootstrapStatuses(); len(got) != len(statuses) {
			t.Fatalf("last statuses should match bootstrap result")
		}
	})

	t.Run("deterministic fallback used when normalized unreachable", func(t *testing.T) {
		cfg := DHTConfig{
			Namespace:      "phase4-test",
			BootstrapPeers: []string{"peer-x", "peer-y"},
		}
		svc, err := NewDHTService(cfg, func(_ context.Context, _ string) bool { return false })
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		statuses := svc.Bootstrap(context.Background())
		normalized := normalizePeers(cfg.BootstrapPeers)
		fallback := deriveFallbackPeers(cfg.Namespace)
		if len(statuses) != len(normalized)+len(fallback) {
			t.Fatalf("expected normalized + fallback statuses, got %d", len(statuses))
		}
		for i, peer := range fallback {
			status := statuses[len(normalized)+i]
			if status.Peer != peer {
				t.Fatalf("expected fallback peer %s, got %s", peer, status.Peer)
			}
			if status.Reason != "deterministic fallback peer unreachable" {
				t.Fatalf("unexpected fallback reason: %s", status.Reason)
			}
		}

		if len(svc.KnownPeers()) != 0 {
			t.Fatalf("expected no known peers after failed bootstrap")
		}
		if got := svc.LastBootstrapStatuses(); len(got) != len(statuses) {
			t.Fatalf("last statuses should mirror bootstrap result")
		}
	})
}

func TestDHTServiceWarmup(t *testing.T) {
	svc := &DHTService{
		cfg:        DHTConfig{WarmupTarget: 3},
		discovered: []string{"peer-1"},
	}
	ctx := context.Background()

	if got := svc.Warmup(ctx); got != svc.cfg.WarmupTarget {
		t.Fatalf("warmup should stop at target, got %d", got)
	}
	if svc.RoutingCount() != svc.cfg.WarmupTarget {
		t.Fatalf("routing count should equal warmup target")
	}
	if got := svc.Warmup(ctx); got != svc.cfg.WarmupTarget {
		t.Fatalf("re-running warmup should not exceed target, got %d", got)
	}

	svc.discovered = nil
	if got := svc.Warmup(ctx); got != svc.cfg.WarmupTarget {
		t.Fatalf("warmup should stay at target when no peers are known, got %d", got)
	}
}

func TestDHTServiceFallbackReasoning(t *testing.T) {
	cfg := DHTConfig{
		Namespace:      "private",
		BootstrapPeers: []string{"peer-x"},
	}
	svc, err := NewDHTService(cfg, func(_ context.Context, _ string) bool { return false })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	status := svc.Bootstrap(context.Background())
	fallback := deriveFallbackPeers(cfg.Namespace)
	if len(status) != len(normalizePeers(cfg.BootstrapPeers))+len(fallback) {
		t.Fatalf("expected fallback statuses to follow normalized ones")
	}
	if len(fallback) == 0 {
		t.Fatal("fallback peers should not be empty")
	}
	fallbackStatuses := status[len(normalizePeers(cfg.BootstrapPeers)):]
	if len(fallbackStatuses) == 0 {
		t.Fatal("fallback peers should not be empty")
	}
	for _, st := range fallbackStatuses {
		if st.Reason != "deterministic fallback peer unreachable" {
			t.Fatalf("unexpected fallback reason for %s: %s", st.Peer, st.Reason)
		}
	}
	if len(svc.KnownPeers()) != 0 {
		t.Fatalf("expected no known peers after fallback, got %v", svc.KnownPeers())
	}
	if got := svc.LastBootstrapStatuses(); len(got) != len(status) {
		t.Fatalf("last statuses should reflect latest bootstrap")
	}
}

func containsPeer(peers []string, peer string) bool {
	for _, p := range peers {
		if p == peer {
			return true
		}
	}
	return false
}
