package phase4

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"
)

type testClock struct {
	now time.Time
}

func (c *testClock) Now() time.Time {
	return c.now
}

func (c *testClock) Advance(d time.Duration) time.Time {
	c.now = c.now.Add(d)
	return c.now
}

func TestMDNSDiscoveryStartupTiming(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, time.February, 15, 10, 0, 0, 0, time.UTC)
	clock := &testClock{now: base}
	d := NewMDNSDiscovery(MDNSConfig{StartupDelay: 5 * time.Minute}, clock.Now)
	if err := d.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if ready := d.StartupReady(); ready {
		t.Fatalf("StartupReady before delay = true")
	}
	clock.Advance(4 * time.Minute)
	if ready := d.StartupReady(); ready {
		t.Fatalf("StartupReady before elapsed delay = true")
	}
	clock.Advance(1 * time.Minute)
	if ready := d.StartupReady(); !ready {
		t.Fatalf("StartupReady after delay = false")
	}
	logs := d.Logs()
	if len(logs) < 2 {
		t.Fatalf("expected at least 2 logs, got %d", len(logs))
	}
	if !strings.Contains(logs[0], "mdns discovery startup started") {
		t.Fatalf("startup log missing, logs=%v", logs)
	}
	if !strings.Contains(logs[1], "mdns discovery startup ready") {
		t.Fatalf("ready log missing, logs=%v", logs)
	}
	prevLen := len(logs)
	if !d.StartupReady() {
		t.Fatalf("StartupReady called again should stay true")
	}
	if got := len(d.Logs()); got != prevLen {
		t.Fatalf("StartupReady logged twice, got %d entries", got)
	}
}

func TestMDNSDiscoveryPeerDedup(t *testing.T) {
	t.Parallel()
	clock := &testClock{now: time.Date(2026, time.February, 15, 10, 0, 0, 0, time.UTC)}
	d := NewMDNSDiscovery(MDNSConfig{}, clock.Now)
	if err := d.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	cases := []struct {
		name string
		peer string
		want bool
	}{
		{name: "first", peer: "peer-a", want: true},
		{name: "trimmed duplicate", peer: " peer-a ", want: false},
		{name: "new peer", peer: "peer-b", want: true},
		{name: "empty", peer: "", want: false},
		{name: "whitespace", peer: "   ", want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := d.AddPeer(tc.peer); got != tc.want {
				t.Fatalf("AddPeer(%q) = %v, want %v", tc.peer, got, tc.want)
			}
		})
	}
	if got := d.KnownPeers(); !equalStringSlices(got, []string{"peer-a", "peer-b"}) {
		t.Fatalf("KnownPeers = %v, want %v", got, []string{"peer-a", "peer-b"})
	}
	logs := d.Logs()
	if !strings.Contains(logs[len(logs)-1], "mdns discovered peer peer-b") {
		t.Fatalf("peer discovery log missing, logs=%v", logs)
	}
}

func TestMDNSDiscoveryDisabled(t *testing.T) {
	t.Parallel()
	clock := &testClock{now: time.Date(2026, time.February, 15, 11, 0, 0, 0, time.UTC)}
	d := NewMDNSDiscovery(MDNSConfig{Disabled: true}, clock.Now)
	if err := d.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !d.IsDisabled() {
		t.Fatalf("IsDisabled = false, want true")
	}
	if ready := d.StartupReady(); ready {
		t.Fatalf("StartupReady when disabled = true")
	}
	if added := d.AddPeer("peer-dis"); added {
		t.Fatalf("AddPeer should be false when disabled")
	}
	logs := d.Logs()
	requireLogContains(t, logs, "mdns discovery disabled for restricted environment")
}

func TestMDNSDiscoveryLogsTimestampedPeers(t *testing.T) {
	t.Parallel()
	clock := &testClock{now: time.Date(2026, time.February, 15, 12, 0, 0, 0, time.UTC)}
	d := NewMDNSDiscovery(MDNSConfig{}, clock.Now)
	if err := d.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if !d.AddPeer("peer-ts") {
		t.Fatalf("AddPeer should succeed")
	}
	var peerLog string
	for _, log := range d.Logs() {
		if strings.Contains(log, "mdns discovered peer peer-ts") {
			peerLog = log
			break
		}
	}
	if peerLog == "" {
		t.Fatalf("peer discovery log not found: %v", d.Logs())
	}
	parts := strings.SplitN(peerLog, " ", 2)
	if len(parts) != 2 {
		t.Fatalf("log entry malformed: %s", peerLog)
	}
	if _, err := time.Parse(time.RFC3339Nano, parts[0]); err != nil {
		t.Fatalf("timestamp parse failed: %v", err)
	}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func requireLogContains(t *testing.T, logs []string, want string) {
	t.Helper()
	for _, log := range logs {
		if strings.Contains(log, want) {
			return
		}
	}
	t.Fatalf("logs do not contain %q: %v", want, logs)
}

func TestMDNSDiscoveryTwoNodeLAN(t *testing.T) {
	t.Parallel()
	lan := newDeterministicLAN(t, []string{"node-a", "node-b"})
	lan.simulateDiscovery()
	lan.assertPeerLists()
	lan.assertDiscoveryLogsTimestamped()
}

type lanNode struct {
	name      string
	discovery *MDNSDiscovery
	clock     *testClock
}

type lanSimulator struct {
	t     *testing.T
	nodes []*lanNode
}

func newDeterministicLAN(t *testing.T, names []string) *lanSimulator {
	t.Helper()
	const startupDelay = time.Second
	base := time.Date(2026, time.February, 15, 13, 0, 0, 0, time.UTC)
	nodes := make([]*lanNode, 0, len(names))
	for i, name := range names {
		clock := &testClock{now: base.Add(time.Duration(i) * time.Minute)}
		d := NewMDNSDiscovery(MDNSConfig{StartupDelay: startupDelay}, clock.Now)
		if err := d.Start(context.Background()); err != nil {
			t.Fatalf("Start(%s): %v", name, err)
		}
		clock.Advance(startupDelay)
		if !d.StartupReady() {
			t.Fatalf("StartupReady(%s) = false", name)
		}
		nodes = append(nodes, &lanNode{name: name, discovery: d, clock: clock})
	}
	return &lanSimulator{t: t, nodes: nodes}
}

func (s *lanSimulator) simulateDiscovery() {
	s.t.Helper()
	const discoveryAdvance = 10 * time.Millisecond
	for _, from := range s.nodes {
		for _, to := range s.nodes {
			if to == from {
				continue
			}
			to.clock.Advance(discoveryAdvance)
			if added := to.discovery.AddPeer(from.name); !added {
				s.t.Fatalf("node %s failed to add peer %s", to.name, from.name)
			}
		}
	}
}

func (s *lanSimulator) assertPeerLists() {
	s.t.Helper()
	for _, node := range s.nodes {
		expected := s.peersFor(node)
		sort.Strings(expected)
		if got := node.discovery.KnownPeers(); !equalStringSlices(got, expected) {
			s.t.Fatalf("node %s KnownPeers = %v, want %v", node.name, got, expected)
		}
	}
}

func (s *lanSimulator) assertDiscoveryLogsTimestamped() {
	s.t.Helper()
	for _, node := range s.nodes {
		for _, peer := range s.peersFor(node) {
			entry := requireLogEntry(s.t, node.discovery.Logs(), fmt.Sprintf("mdns discovered peer %s", peer))
			assertLogTimestamp(s.t, entry)
			s.t.Logf("lan-discovery node=%s entry=%s", node.name, entry)
		}
	}
}

func (s *lanSimulator) peersFor(node *lanNode) []string {
	peers := make([]string, 0, len(s.nodes)-1)
	for _, other := range s.nodes {
		if other == node {
			continue
		}
		peers = append(peers, other.name)
	}
	return peers
}

func requireLogEntry(t *testing.T, logs []string, want string) string {
	t.Helper()
	for _, log := range logs {
		if strings.Contains(log, want) {
			return log
		}
	}
	t.Fatalf("log entry %q not found: %v", want, logs)
	return ""
}

func assertLogTimestamp(t *testing.T, entry string) {
	t.Helper()
	parts := strings.SplitN(entry, " ", 2)
	if len(parts) != 2 {
		t.Fatalf("log entry malformed: %s", entry)
	}
	if _, err := time.Parse(time.RFC3339Nano, parts[0]); err != nil {
		t.Fatalf("timestamp parse failed: %v", err)
	}
}
