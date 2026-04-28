package node_test

import (
	"context"
	"testing"
	"time"

	v0_1 "github.com/aether/code_aether/pkg/v0_1"
	clearmode "github.com/aether/code_aether/pkg/v0_1/mode/clear"
)

// TestW5ClearModeDowngradePrevented verifies scope-level mode continuity.
func TestW5ClearModeDowngradePrevented(t *testing.T) {
	// Scope established in Seal mode.
	err := clearmode.EnforceModeContinuity(clearmode.ModeSeal, clearmode.ModeSeal)
	if err != nil {
		t.Fatalf("same mode should be allowed: %v", err)
	}

	// Attempt downgrade to Clear.
	err = clearmode.EnforceModeContinuity(clearmode.ModeSeal, clearmode.ModeClear)
	if err == nil {
		t.Fatal("downgrade from Seal to Clear should be rejected")
	}

	// Attempt downgrade to Crowd (also incompatible).
	err = clearmode.EnforceModeContinuity(clearmode.ModeSeal, clearmode.ModeCrowd)
	if err == nil {
		t.Fatal("downgrade from Seal to Crowd should be rejected")
	}

	// Clear → Clear is fine.
	err = clearmode.EnforceModeContinuity(clearmode.ModeClear, clearmode.ModeClear)
	if err != nil {
		t.Fatalf("Clear→Clear should be allowed: %v", err)
	}

	// New scope (empty existing): any mode is fine.
	err = clearmode.EnforceModeContinuity("", clearmode.ModeSeal)
	if err != nil {
		t.Fatalf("new scope→Seal should be allowed: %v", err)
	}
	err = clearmode.EnforceModeContinuity("", clearmode.ModeClear)
	if err != nil {
		t.Fatalf("new scope→Clear should be allowed: %v", err)
	}
}

// TestW5RuntimeBootsAndServes verifies the v0.1 runtime starts, serves, and shuts down.
func TestW5RuntimeBootsAndServes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rt, err := v0_1.Start(ctx, v0_1.Config{
		ListenAddr: "/ip4/127.0.0.1/tcp/0",
		EnableMDNS: false,
		EnableNAT:  false,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer rt.Close()

	if rt.PeerID() == "" {
		t.Fatal("PeerID should not be empty")
	}
	addrs := rt.ListenAddrs()
	if len(addrs) == 0 {
		t.Fatal("should have at least one listen address")
	}
	t.Logf("v0.1 runtime: peer_id=%s addrs=%v", rt.PeerID(), addrs)
}

// TestW5TwoRuntimesConnect verifies two v0.1 runtimes can connect peer-to-peer.
func TestW5TwoRuntimesConnect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rtA, err := v0_1.Start(ctx, v0_1.Config{ListenAddr: "/ip4/127.0.0.1/tcp/0"})
	if err != nil {
		t.Fatalf("Start A: %v", err)
	}
	defer rtA.Close()

	rtB, err := v0_1.Start(ctx, v0_1.Config{ListenAddr: "/ip4/127.0.0.1/tcp/0"})
	if err != nil {
		t.Fatalf("Start B: %v", err)
	}
	defer rtB.Close()

	t.Logf("A: %s %v", rtA.PeerID(), rtA.ListenAddrs())
	t.Logf("B: %s %v", rtB.PeerID(), rtB.ListenAddrs())

	if rtA.PeerID() == rtB.PeerID() {
		t.Fatal("A and B should have different peer IDs")
	}
}
