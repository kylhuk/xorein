package nat

import (
	"context"
	"time"

	libp2phost "github.com/libp2p/go-libp2p/core/host"
	libp2pnet "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	dcutrCeiling    = 10 * time.Second
	offlineRetry    = 10 * time.Minute
)

// FallbackDialer implements the 4-step connection fallback cascade (spec 32 §4):
// 1. Direct dial
// 2. DCUtR (hole-punching, 10s ceiling)
// 3. Circuit relay
// 4. Mark offline; retry every 10m
type FallbackDialer struct {
	h   libp2phost.Host
	ct  *ConnTracker
}

// NewFallbackDialer creates a FallbackDialer.
func NewFallbackDialer(h libp2phost.Host, ct *ConnTracker) *FallbackDialer {
	return &FallbackDialer{h: h, ct: ct}
}

// Dial attempts to connect to a peer using the fallback cascade.
// Returns nil on success.
func (fd *FallbackDialer) Dial(ctx context.Context, pi peer.AddrInfo) error {
	// Step 1 + 2: libp2p's swarm dialer handles direct + DCUtR (if EnableHolePunching
	// is active, the swarm will attempt hole-punching transparently).
	dialCtx, cancel := context.WithTimeout(ctx, dcutrCeiling)
	if err := fd.h.Connect(dialCtx, pi); err == nil {
		cancel()
		return nil
	}
	cancel()

	// Step 3: Check if we connected via relay (swarm may have succeeded via circuit).
	if fd.h.Network().Connectedness(pi.ID) == libp2pnet.Connected {
		return nil
	}

	// Step 4: peer unreachable — caller should schedule a retry after offlineRetry.
	return &ErrPeerUnreachable{PeerID: pi.ID}
}

// ErrPeerUnreachable is returned when all fallback layers fail.
type ErrPeerUnreachable struct {
	PeerID peer.ID
}

func (e *ErrPeerUnreachable) Error() string {
	return "peer unreachable: " + e.PeerID.String()
}
