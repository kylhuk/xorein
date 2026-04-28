// Package v0_1 provides the v0.1 runtime entry point.
package v0_1

import (
	"context"
	"fmt"

	libp2p "github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	libp2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	holepunch "github.com/libp2p/go-libp2p/p2p/protocol/holepunch"

	apb "github.com/aether/code_aether/gen/go/proto"
	chatpkg "github.com/aether/code_aether/pkg/v0_1/family/chat"
	dmpkg "github.com/aether/code_aether/pkg/v0_1/family/dm"
	friendspkg "github.com/aether/code_aether/pkg/v0_1/family/friends"
	groupdmpkg "github.com/aether/code_aether/pkg/v0_1/family/groupdm"
	governancepkg "github.com/aether/code_aether/pkg/v0_1/family/governance"
	identitypkg "github.com/aether/code_aether/pkg/v0_1/family/identity"
	idstore "github.com/aether/code_aether/pkg/v0_1/family/identity/store"
	manifestpkg "github.com/aether/code_aether/pkg/v0_1/family/manifest"
	moderationpkg "github.com/aether/code_aether/pkg/v0_1/family/moderation"
	notifypkg "github.com/aether/code_aether/pkg/v0_1/family/notify"
	peerfamily "github.com/aether/code_aether/pkg/v0_1/family/peer"
	presencepkg "github.com/aether/code_aether/pkg/v0_1/family/presence"
	syncpkg "github.com/aether/code_aether/pkg/v0_1/family/sync"
	voicepkg "github.com/aether/code_aether/pkg/v0_1/family/voice"
	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
	"github.com/aether/code_aether/pkg/v0_1/control"
	"github.com/aether/code_aether/pkg/v0_1/discovery"
	"github.com/aether/code_aether/pkg/v0_1/nat"
	"github.com/aether/code_aether/pkg/v0_1/transport"
)

// Config is the v0.1 runtime configuration.
type Config struct {
	// ListenAddr is the libp2p TCP listen address, e.g. "/ip4/127.0.0.1/tcp/0".
	ListenAddr string
	// Identity is this node's signed identity profile (public key metadata).
	Identity *apb.IdentityProfile
	// PrivateKey is the Ed25519 private key used to derive the libp2p PeerID (spec 30 §2.1).
	// If nil a new ephemeral key is generated each run.
	PrivateKey libp2pcrypto.PrivKey
	// Capabilities is the list of advertised feature flags.
	Capabilities []proto.FeatureFlag
	// Role determines which family handlers are registered (spec 30 §1.4).
	// Valid values: "client" (default), "relay", "bootstrap", "archivist".
	Role string
	// EnableMDNS enables LAN peer discovery via mDNS.
	EnableMDNS bool
	// EnableNAT enables NAT port mapping and hole punching.
	EnableNAT bool
	// DataDir is the path to the data directory (for state.db, control.token, etc.).
	// If empty, the control API server is not started.
	DataDir string
	// ControlAddr overrides the default control socket/addr path.
	ControlAddr string
	// BootstrapAddrs is a list of "/ip4/.../tcp/.../p2p/<id>" multiaddrs for bootstrap nodes.
	BootstrapAddrs []string
	// ManualPeers is a list of peer multiaddrs or host:port addresses to always connect to.
	ManualPeers []string
	// RelayAddrs is a list of circuit-relay v2 server multiaddrs (for auto-relay client).
	RelayAddrs []string
	// RelayListenAddr is an optional second TCP listen address for relay-role nodes (spec 30 §5).
	// If empty, relay nodes share the primary ListenAddr.
	RelayListenAddr string
}

// Runtime holds a running v0.1 libp2p host with all registered family handlers.
type Runtime struct {
	host         libp2phost.Host
	cfg          Config
	peerCache    *discovery.Cache
	mdnsSvc      *discovery.MDNSService
	dhtSvc       *discovery.DHT
	discoverLoop *discovery.Loop
	relaySvc     *nat.RelayService
	connTracker  *nat.ConnTracker
	dcutrTracer  *nat.DCUtRTracer
	controlSrv   *control.Server
	cancelDisc   context.CancelFunc
}

// Start creates and starts the v0.1 libp2p host and registers protocol family handlers.
func Start(ctx context.Context, cfg Config) (*Runtime, error) {
	listenAddr := cfg.ListenAddr
	if listenAddr == "" {
		listenAddr = "/ip4/127.0.0.1/tcp/0"
	}

	cache := discovery.NewCache()
	rt := &Runtime{cfg: cfg, peerCache: cache}

	// Create the DCUtR tracer early so we can pass it to EnableHolePunching.
	// The ConnTracker reference is set after the host is built.
	rt.dcutrTracer = &nat.DCUtRTracer{}

	// Validate role capabilities (spec 03 §3.4). Use role's cap set when none supplied.
	if len(cfg.Capabilities) == 0 {
		cfg.Capabilities = proto.RoleCapabilities(cfg.Role)
	}

	opts := transport.StandardOptions(cfg.PrivateKey)
	listenAddrs := []string{listenAddr}
	if cfg.RelayListenAddr != "" && cfg.Role == "relay" {
		listenAddrs = append(listenAddrs, cfg.RelayListenAddr)
	}
	opts = append(opts, libp2p.ListenAddrStrings(listenAddrs...))

	if cfg.EnableNAT {
		opts = append(opts,
			libp2p.NATPortMap(),
			libp2p.EnableHolePunching(holepunch.WithTracer(rt.dcutrTracer)),
			libp2p.EnableNATService(),
		)
	}

	if len(cfg.RelayAddrs) > 0 {
		relayPeers := nat.ParseRelayAddrs(cfg.RelayAddrs)
		if len(relayPeers) > 0 {
			opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(relayPeers))
		}
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("v0.1 runtime: libp2p host: %w", err)
	}
	rt.host = h

	// Connection type tracker — wires the DCUtR tracer to upgrade labels.
	rt.connTracker = nat.NewConnTracker(h.Network())
	rt.dcutrTracer.SetConnTracker(rt.connTracker)

	// Register protocol family handlers.
	hs := rt.registerFamilies()

	// Relay service for relay-role nodes.
	if rt.role() == "relay" {
		rs, err := nat.NewRelayService(h)
		if err != nil {
			h.Close()
			return nil, fmt.Errorf("v0.1 runtime: relay service: %w", err)
		}
		rt.relaySvc = rs
	}

	// Discovery stack.
	if err := rt.startDiscovery(ctx, cache, hs); err != nil {
		h.Close()
		return nil, err
	}

	// Wire ConnectionTypeFn so the control API can report connection types.
	hs.ConnectionTypeFn = rt.ConnectionType
	// Wire LatencyFn so the control API can report EWMA latency (spec 32 §5).
	hs.LatencyFn = rt.PeerLatencyMs

	// Control API.
	if cfg.DataDir != "" {
		srv, err := control.New(control.Config{
			DataDir:  cfg.DataDir,
			Addr:     cfg.ControlAddr,
			Handlers: hs,
		})
		if err != nil {
			h.Close()
			return nil, fmt.Errorf("v0.1 runtime: control API: %w", err)
		}
		rt.controlSrv = srv
		go srv.Serve() //nolint:errcheck
	}

	return rt, nil
}

// startDiscovery initialises and launches all spec 31 discovery layers.
func (rt *Runtime) startDiscovery(ctx context.Context, cache *discovery.Cache, _ *control.Handlers) error {
	h := rt.host
	role := rt.role()
	caps := make([]string, len(rt.cfg.Capabilities))
	for i, c := range rt.cfg.Capabilities {
		caps[i] = string(c)
	}

	manual := discovery.NewManualPeers(rt.cfg.ManualPeers)

	bc, err := discovery.NewBootstrapClient(h, rt.cfg.BootstrapAddrs, cache)
	if err != nil {
		return fmt.Errorf("v0.1 runtime: bootstrap client: %w", err)
	}
	// Bootstrap role serves bootstrap.fetch/register for other nodes.
	if role == "bootstrap" {
		discovery.NewBootstrapServer(h, cache)
	}

	var dhtSvc *discovery.DHT
	if role != "bootstrap" {
		d, err := discovery.NewDHT(ctx, h, cache, bc.BootstrapAddrs())
		if err == nil {
			dhtSvc = d
		}
		rt.dhtSvc = dhtSvc
	}

	pex := discovery.NewPEX(h, cache)

	var mdnsSvc *discovery.MDNSService
	if rt.cfg.EnableMDNS {
		mdnsSvc = discovery.NewMDNSService(h, role, caps, cache)
		if err := mdnsSvc.Start(ctx); err == nil {
			rt.mdnsSvc = mdnsSvc
		} else {
			mdnsSvc = nil
		}
	}

	discCtx, cancel := context.WithCancel(ctx)
	rt.cancelDisc = cancel

	loop := discovery.NewLoop(discovery.LoopConfig{
		Host:      h,
		Cache:     cache,
		MDNS:      mdnsSvc,
		DHT:       dhtSvc,
		Bootstrap: bc,
		PEX:       pex,
		Manual:    manual,
	})
	rt.discoverLoop = loop
	go loop.Run(discCtx)

	return nil
}

// role returns the normalized role string, defaulting to "client".
func (rt *Runtime) role() string {
	if rt.cfg.Role == "" {
		return "client"
	}
	return rt.cfg.Role
}

// registerFamilies wires protocol family handlers per spec 30 §1.4 role constraints.
// It returns a Handlers struct for the control API.
func (rt *Runtime) registerFamilies() *control.Handlers {
	reg := transport.RegisterFamilyWithDeadlines
	localPeer := rt.host.ID().String()

	hs := &control.Handlers{PeerID: localPeer}

	// peer family: all roles.
	peerHandler := &peerfamily.Handler{
		LocalIdentity: rt.cfg.Identity,
		LocalCaps:     rt.cfg.Capabilities,
	}
	reg(rt.host, peerHandler)
	hs.Peer = peerHandler

	// bootstrap role: peer family only (spec 30 §1.4).
	if rt.role() == "bootstrap" {
		return hs
	}

	// identity: relay + client + archivist.
	identHandler := &identitypkg.Handler{
		BundleStore: idstore.NewMemStore(),
	}
	reg(rt.host, identHandler)
	hs.Identity = identHandler

	// relay role: peer + identity only (spec 30 §1.4 — no chat/DM).
	if rt.role() == "relay" {
		return hs
	}

	// All remaining families: client + archivist.

	dmHandler := &dmpkg.Handler{
		LocalPeerID: localPeer,
		BundleStore: identHandler.BundleStore,
	}
	reg(rt.host, dmHandler)
	hs.DM = dmHandler

	groupdmHandler := &groupdmpkg.Handler{LocalPeerID: localPeer}
	reg(rt.host, groupdmHandler)
	hs.GroupDM = groupdmHandler

	voiceHandler := &voicepkg.Handler{LocalPeerID: localPeer}
	reg(rt.host, voiceHandler)
	hs.Voice = voiceHandler

	govHandler := governancepkg.NewHandler()
	reg(rt.host, govHandler)
	hs.Governance = govHandler

	modHandler := moderationpkg.New(govHandler)
	reg(rt.host, modHandler)
	hs.Moderation = modHandler

	chatH := &chatpkg.Handler{}
	chatH.WithModeration(modHandler)
	reg(rt.host, chatH)
	hs.Chat = chatH

	manifestH := &manifestpkg.Handler{}
	reg(rt.host, manifestH)
	hs.Manifest = manifestH

	presenceH := &presencepkg.Handler{}
	reg(rt.host, presenceH)
	hs.Presence = presenceH

	notifyH := &notifypkg.Handler{}
	reg(rt.host, notifyH)
	hs.Notify = notifyH

	friendsH := &friendspkg.Handler{LocalPeerID: localPeer}
	reg(rt.host, friendsH)
	hs.Friends = friendsH

	// archivist: sync handler with full-coverage guarantees.
	isArchivist := rt.role() == "archivist"
	syncH := syncpkg.NewHandler(isArchivist)
	reg(rt.host, syncH)
	hs.Sync = syncH

	return hs
}

// ListenAddrs returns the multiaddresses this host is listening on.
func (rt *Runtime) ListenAddrs() []string {
	addrs := rt.host.Addrs()
	strs := make([]string, len(addrs))
	for i, a := range addrs {
		strs[i] = a.String()
	}
	return strs
}

// PeerID returns the libp2p peer ID string for this host.
func (rt *Runtime) PeerID() string {
	return rt.host.ID().String()
}

// ControlAddr returns the control API socket/addr string, or empty if not started.
func (rt *Runtime) ControlAddr() string {
	if rt.controlSrv == nil {
		return ""
	}
	return rt.controlSrv.Addr()
}

// ConnectionType returns the detected connection type for a given peer ID string.
func (rt *Runtime) ConnectionType(peerIDStr string) string {
	if rt.connTracker == nil {
		return string(nat.ConnUnknown)
	}
	pid, err := peer.Decode(peerIDStr)
	if err != nil {
		return string(nat.ConnUnknown)
	}
	return string(rt.connTracker.Get(pid))
}

// PeerLatencyMs returns the EWMA latency in milliseconds to the given peer,
// or -1 if the peer is not tracked or the peerstore does not hold latency data.
// Used to populate the latency_ms field on the /v1/peers/{peerID}/connection
// control API endpoint (spec 32 §5).
func (rt *Runtime) PeerLatencyMs(peerIDStr string) int64 {
	pid, err := peer.Decode(peerIDStr)
	if err != nil {
		return -1
	}
	d := rt.host.Peerstore().LatencyEWMA(pid)
	if d == 0 {
		return -1
	}
	return d.Milliseconds()
}

// Close shuts down the libp2p host and any discovery or control services.
func (rt *Runtime) Close() error {
	if rt.cancelDisc != nil {
		rt.cancelDisc()
	}
	if rt.controlSrv != nil {
		rt.controlSrv.Shutdown(context.Background()) //nolint:errcheck
	}
	if rt.mdnsSvc != nil {
		rt.mdnsSvc.Close()
	}
	if rt.dhtSvc != nil {
		rt.dhtSvc.Close() //nolint:errcheck
	}
	if rt.relaySvc != nil {
		rt.relaySvc.Close() //nolint:errcheck
	}
	return rt.host.Close()
}
