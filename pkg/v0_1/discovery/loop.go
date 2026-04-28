package discovery

import (
	"context"
	"time"

	libp2phost "github.com/libp2p/go-libp2p/core/host"
	libp2pnet "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	discoveryInterval        = 250 * time.Millisecond
	dhtWalkInterval          = 10 * time.Second // gate DHT walks to avoid hammering bootstrap
	defaultTargetConnections = 5
)

// Loop runs the layered peer discovery loop (spec 31 §1, 7 layers).
type Loop struct {
	h               libp2phost.Host
	cache           *Cache
	mdns            *MDNSService
	dht             *DHT
	bootstrap       *BootstrapClient
	pex             *PEX
	manual          *ManualPeers
	targetConns     int
	lastDHTWalk     time.Time
	lastBootstrap   time.Time
	bootstrapInterval time.Duration
}

// LoopConfig holds options for the discovery loop.
type LoopConfig struct {
	Host            libp2phost.Host
	Cache           *Cache
	MDNS            *MDNSService      // may be nil
	DHT             *DHT              // may be nil
	Bootstrap       *BootstrapClient  // may be nil
	PEX             *PEX              // may be nil
	Manual          *ManualPeers      // may be nil
	TargetConns     int               // 0 defaults to defaultTargetConnections
}

// NewLoop creates a discovery Loop.
func NewLoop(cfg LoopConfig) *Loop {
	target := cfg.TargetConns
	if target < defaultTargetConnections {
		target = defaultTargetConnections
	}
	return &Loop{
		h:                 cfg.Host,
		cache:             cfg.Cache,
		mdns:              cfg.MDNS,
		dht:               cfg.DHT,
		bootstrap:         cfg.Bootstrap,
		pex:               cfg.PEX,
		manual:            cfg.Manual,
		targetConns:       target,
		bootstrapInterval: 30 * time.Second,
	}
}

// Run starts the discovery loop and blocks until ctx is cancelled.
func (l *Loop) Run(ctx context.Context) {
	// Seed cache from manual peers immediately.
	if l.manual != nil {
		l.manual.SeedCache(l.cache)
	}

	// Initial bootstrap fetch.
	if l.bootstrap != nil {
		go l.bootstrap.RegisterSelf(ctx)
		go l.bootstrap.FetchPeers(ctx)
	}

	ticker := time.NewTicker(discoveryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.tick(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (l *Loop) tick(ctx context.Context) {
	connected := len(l.h.Network().Peers())
	if connected >= l.targetConns {
		return
	}

	now := time.Now()

	// Layer 1: attempt reconnect to cache entries ready for retry.
	l.reconnectFromCache(ctx)

	// Layer 3: DHT walk (rate-limited to every dhtWalkInterval).
	if l.dht != nil && now.Sub(l.lastDHTWalk) >= dhtWalkInterval {
		l.lastDHTWalk = now
		go l.dht.FindProviders(ctx)
		go func() {
			_ = l.dht.AnnounceProvider(ctx)
		}()
	}

	// Layer 4: bootstrap fetch (rate-limited).
	if l.bootstrap != nil && now.Sub(l.lastBootstrap) >= l.bootstrapInterval {
		l.lastBootstrap = now
		go l.bootstrap.FetchPeers(ctx)
	}

	// Layer 6: PEX with one random connected peer.
	if l.pex != nil && connected > 0 {
		go l.pex.ExchangeWith(ctx)
	}
}

func (l *Loop) reconnectFromCache(ctx context.Context) {
	records := l.cache.All()
	for _, r := range records {
		if r.PeerID == "" || r.PeerID == l.h.ID().String() {
			continue
		}
		// Skip peers already connected.
		pid, err := peer.Decode(r.PeerID)
		if err != nil {
			continue
		}
		if l.h.Network().Connectedness(pid) == libp2pnet.Connected {
			continue
		}
		if !l.cache.ReadyForRetry(r.PeerID) {
			continue
		}

		addrs := make([]ma.Multiaddr, 0, len(r.Addresses))
		for _, a := range r.Addresses {
			maddr, err := ma.NewMultiaddr(a)
			if err == nil {
				addrs = append(addrs, maddr)
			}
		}
		if len(addrs) == 0 {
			continue
		}

		pi := peer.AddrInfo{ID: pid, Addrs: addrs}
		go func(pi peer.AddrInfo, peerID string) {
			if err := l.h.Connect(ctx, pi); err != nil {
				l.cache.RecordFailure(peerID)
			}
		}(pi, r.PeerID)
	}
}
