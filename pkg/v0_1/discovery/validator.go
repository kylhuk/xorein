package discovery

import (
	"context"

	libp2ppeer "github.com/libp2p/go-libp2p/core/peer"
)

// ValidatePeer verifies a peer by performing a peer.info round-trip.
// This stub marks peers as tentative; full implementation requires a live host.
// Returns nil to allow the peer through (tentative cache entry).
func ValidatePeer(ctx context.Context, info libp2ppeer.AddrInfo) error {
	// TODO: open Noise stream, call peer.info, verify hybrid signature on response.
	_ = ctx
	_ = info
	return nil
}
