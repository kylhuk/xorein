package discovery

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	libp2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pprotocol "github.com/libp2p/go-libp2p/core/protocol"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	dhtProtocolID  = "/aether/kad/0.1.0"
	dhtProviderTTL = 24 * time.Hour
	dhtReannounce  = 22 * time.Hour
)

// DHT wraps a kad-dht instance with provider record management.
type DHT struct {
	h     libp2phost.Host
	d     *dht.IpfsDHT
	cache *Cache
}

// NewDHT creates and bootstraps a kad-dht node with spec 31 §3.1 protocol ID.
func NewDHT(ctx context.Context, h libp2phost.Host, cache *Cache, bootstrapPeers []peer.AddrInfo) (*DHT, error) {
	d, err := dht.New(ctx, h,
		dht.V1ProtocolOverride(libp2pprotocol.ID(dhtProtocolID)),
		dht.Mode(dht.ModeAuto),
	)
	if err != nil {
		return nil, fmt.Errorf("dht: new: %w", err)
	}

	for _, bp := range bootstrapPeers {
		_ = h.Connect(ctx, bp) // best-effort; routing table fills over time
	}

	if err := d.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("dht: bootstrap: %w", err)
	}

	return &DHT{h: h, d: d, cache: cache}, nil
}

// AnnounceProvider stores a provider record for this node under SHA-256(peerID).
func (dd *DHT) AnnounceProvider(ctx context.Context) error {
	c, err := peerCID(dd.h.ID().String())
	if err != nil {
		return err
	}
	return dd.d.Provide(ctx, c, true)
}

// FindProviders discovers peers advertising the same DHT key and loads them into the cache.
func (dd *DHT) FindProviders(ctx context.Context) {
	c, err := peerCID(dd.h.ID().String())
	if err != nil {
		return
	}
	fCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	ch := dd.d.FindProvidersAsync(fCtx, c, 20)
	for pi := range ch {
		if pi.ID == dd.h.ID() {
			continue
		}
		addrs := make([]string, len(pi.Addrs))
		for i, a := range pi.Addrs {
			addrs[i] = a.String()
		}
		r := PeerRecord{
			PeerID:    pi.ID.String(),
			Addresses: addrs,
			Source:    "dht",
		}
		dd.cache.Put(r, addrTTL)
		dd.h.Peerstore().AddAddrs(pi.ID, pi.Addrs, dhtProviderTTL)
	}
}

// FindPeer resolves addresses for a known peer ID via the DHT.
func (dd *DHT) FindPeer(ctx context.Context, id peer.ID) ([]ma.Multiaddr, error) {
	fCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	pi, err := dd.d.FindPeer(fCtx, id)
	if err != nil {
		return nil, err
	}
	return pi.Addrs, nil
}

// Close shuts down the DHT.
func (dd *DHT) Close() error {
	return dd.d.Close()
}

// peerCID builds a CIDv1 with raw SHA-256 of the peer ID string.
func peerCID(peerID string) (cid.Cid, error) {
	sum := sha256.Sum256([]byte(peerID))
	mh, err := multihash.Encode(sum[:], multihash.SHA2_256)
	if err != nil {
		return cid.Undef, fmt.Errorf("dht: multihash encode: %w", err)
	}
	return cid.NewCidV1(cid.Raw, mh), nil
}
