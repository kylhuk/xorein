// Package nat implements NAT traversal for the xorein v0.1 runtime (spec 32).
package nat

import (
	"strings"
	"sync"

	libp2pnet "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

// ConnectionType classifies how a peer is reachable.
type ConnectionType string

const (
	ConnDirect      ConnectionType = "direct"
	ConnDCUtR       ConnectionType = "dcutr"
	ConnCircuitV2   ConnectionType = "circuit-relay-v2"
	ConnXoreinRelay ConnectionType = "xorein-relay"
	ConnUnknown     ConnectionType = "unknown"
)

// ConnTracker tracks per-peer connection types using libp2p network notifications.
type ConnTracker struct {
	mu    sync.RWMutex
	types map[peer.ID]ConnectionType
}

var _ libp2pnet.Notifiee = (*ConnTracker)(nil)

// NewConnTracker creates a ConnTracker and registers it with the host's network.
func NewConnTracker(n libp2pnet.Network) *ConnTracker {
	ct := &ConnTracker{types: make(map[peer.ID]ConnectionType)}
	n.Notify(ct)
	return ct
}

// Get returns the connection type for a peer, or ConnUnknown if not tracked.
func (ct *ConnTracker) Get(id peer.ID) ConnectionType {
	ct.mu.RLock()
	t, ok := ct.types[id]
	ct.mu.RUnlock()
	if !ok {
		return ConnUnknown
	}
	return t
}

// Listen implements network.Notifiee.
func (ct *ConnTracker) Listen(libp2pnet.Network, ma.Multiaddr) {}

// ListenClose implements network.Notifiee.
func (ct *ConnTracker) ListenClose(libp2pnet.Network, ma.Multiaddr) {}

// Connected implements network.Notifiee — classifies the connection on open.
func (ct *ConnTracker) Connected(_ libp2pnet.Network, conn libp2pnet.Conn) {
	ct.mu.Lock()
	ct.types[conn.RemotePeer()] = classifyConn(conn)
	ct.mu.Unlock()
}

// Disconnected implements network.Notifiee — removes the peer entry on close.
func (ct *ConnTracker) Disconnected(_ libp2pnet.Network, conn libp2pnet.Conn) {
	ct.mu.Lock()
	delete(ct.types, conn.RemotePeer())
	ct.mu.Unlock()
}

// classifyConn determines the connection type from the remote multiaddr.
func classifyConn(conn libp2pnet.Conn) ConnectionType {
	if conn.Stat().Limited {
		// Limited == true means this is a relayed connection.
		remote := conn.RemoteMultiaddr().String()
		if strings.Contains(remote, "p2p-circuit") {
			return ConnCircuitV2
		}
		return ConnXoreinRelay
	}
	// Direct connections: check if upgraded via hole-punch (DCUtR).
	// DCUtR-upgraded connections appear as direct but were originally relayed.
	// We mark them as direct here; the DCUtR tracer upgrades the label if needed.
	return ConnDirect
}

// MarkDCUtR upgrades a peer's connection type to dcutr after a successful hole-punch.
func (ct *ConnTracker) MarkDCUtR(id peer.ID) {
	ct.mu.Lock()
	ct.types[id] = ConnDCUtR
	ct.mu.Unlock()
}
