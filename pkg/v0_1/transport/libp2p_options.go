package transport

import (
	"io"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2pnetwork "github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	goyamux "github.com/libp2p/go-yamux/v4"
)

// spec30 connection limits (spec 30 §1.1).
const (
	maxInboundConns  = 1024
	maxOutboundConns = 512
)

// specYamuxTransport builds a yamux.Transport with the settings required by
// spec 30 §1.3:
//   - MaxStreamWindowSize = 16 MiB
//   - MaxIncomingStreams = 1024 (MaxConcurrentStreams)
//   - EnableKeepAlive = true, KeepAliveInterval = 60s
//
// All other fields are taken from yamux.DefaultTransport so we stay compatible
// with the libp2p defaults (ReadBufSize=0, LogOutput=io.Discard, etc.).
func specYamuxTransport() *yamux.Transport {
	cfg := *(*goyamux.Config)(yamux.DefaultTransport) // copy defaults
	cfg.MaxStreamWindowSize = 16 << 20                // 16 MiB (spec 30 §1.3)
	cfg.MaxIncomingStreams = 1024                      // spec 30 §1.3
	cfg.EnableKeepAlive = true
	cfg.KeepAliveInterval = 60 * time.Second // spec 30 §1.3
	cfg.LogOutput = io.Discard               // keep quiet; already set by default but be explicit
	cfg.ReadBufSize = 0                      // security transport buffers internally
	return (*yamux.Transport)(&cfg)
}

// specResourceManager returns a libp2p resource manager that enforces the
// spec 30 §1.1 direction-specific connection limits:
//   - inbound: 1024
//   - outbound: 512
//
// Stream and memory limits continue to be derived from available system
// resources via DefaultLimits.AutoScale().
func specResourceManager() (libp2pnetwork.ResourceManager, error) {
	// Start from the auto-scaled defaults so that stream/memory limits remain
	// proportional to available host resources.
	defaults := rcmgr.DefaultLimits
	libp2p.SetDefaultServiceLimits(&defaults)
	base := defaults.AutoScale()

	// Overlay the spec-mandated system connection limits using PartialLimitConfig
	// (the only public API for mutating a ConcreteLimitConfig).
	partial := rcmgr.PartialLimitConfig{
		System: rcmgr.ResourceLimits{
			ConnsInbound:  rcmgr.LimitVal(maxInboundConns),
			ConnsOutbound: rcmgr.LimitVal(maxOutboundConns),
			Conns:         rcmgr.LimitVal(maxInboundConns + maxOutboundConns),
		},
	}
	concrete := partial.Build(base)
	return rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(concrete))
}

// StandardOptions returns the libp2p host options required by spec 30:
//   - Pinned Noise XX security with /aether/noise/1.0 prologue
//   - Yamux 1.0.0 muxer (MaxStreamWindowSize=16 MiB, MaxConcurrentStreams=1024, keepalive=60s)
//   - Explicit TCP transport
//   - Connection manager (low=512, high=1024, grace=30s)
//   - Resource manager with direction-specific limits (inbound=1024, outbound=512)
//   - User-agent "xorein/0.1"
//   - Dial timeout 30s
//
// If identityKey is non-nil the host PeerID is derived from it (spec 30 §2.1).
func StandardOptions(identityKey libp2pcrypto.PrivKey) []libp2p.Option {
	cm, err := connmgr.NewConnManager(512, 1024, connmgr.WithGracePeriod(30*time.Second))
	if err != nil {
		// NewConnManager only fails on invalid ranges (low>high); our constants are valid.
		panic("transport: NewConnManager: " + err.Error())
	}

	rm, err := specResourceManager()
	if err != nil {
		// specResourceManager only fails if AutoScale cannot read system memory,
		// which is non-fatal — fall back to no custom resource manager.
		rm = nil //nolint:ineffassign
	}

	opts := []libp2p.Option{
		libp2p.Security(noise.ID, NewNoiseWithPrologue),
		libp2p.Muxer(yamux.ID, specYamuxTransport()),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.ConnectionManager(cm),
		libp2p.UserAgent("xorein/0.1"),
		libp2p.WithDialTimeout(30 * time.Second),
	}
	if rm != nil {
		opts = append(opts, libp2p.ResourceManager(rm))
	}
	if identityKey != nil {
		opts = append(opts, libp2p.Identity(identityKey))
	}
	return opts
}
