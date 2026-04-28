package transport

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/multiformats/go-multiaddr"
)

const (
	// BootstrapAddressTTL is the peerstore TTL for bootstrap-learned addresses (spec 30 §4.1).
	BootstrapAddressTTL = 24 * time.Hour
	// MDNSAddressTTL is the peerstore TTL for mDNS-learned addresses (spec 30 §4.1).
	MDNSAddressTTL = 15 * time.Minute
)

// AddBootstrapAddr adds addr to the peerstore with BootstrapAddressTTL (spec 30 §4.1).
func AddBootstrapAddr(ps peerstore.Peerstore, id peer.ID, addr multiaddr.Multiaddr) {
	ps.AddAddr(id, addr, BootstrapAddressTTL)
}

// AddMDNSAddr adds addr to the peerstore with MDNSAddressTTL (spec 30 §4.1).
func AddMDNSAddr(ps peerstore.Peerstore, id peer.ID, addr multiaddr.Multiaddr) {
	ps.AddAddr(id, addr, MDNSAddressTTL)
}
