package discovery

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	libp2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/network"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	bootstrapProtocol = "/aether/bootstrap/0.1.0"
	bootstrapTimeout  = 10 * time.Second
	bootstrapMaxPeers = 200
)

// BootstrapClient fetches peers from one or more bootstrap nodes.
type BootstrapClient struct {
	h     libp2phost.Host
	peers []peer.AddrInfo
	cache *Cache
}

// NewBootstrapClient creates a BootstrapClient targeting the given bootstrap peers.
func NewBootstrapClient(h libp2phost.Host, bootstrapAddrs []string, cache *Cache) (*BootstrapClient, error) {
	var pis []peer.AddrInfo
	for _, addr := range bootstrapAddrs {
		addr = strings.TrimSpace(addr)
		if addr == "" {
			continue
		}
		maddr, err := ma.NewMultiaddr(addr)
		if err != nil {
			continue
		}
		pi, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			continue
		}
		pis = append(pis, *pi)
	}
	return &BootstrapClient{h: h, peers: pis, cache: cache}, nil
}

// BootstrapAddrs returns the parsed bootstrap peer.AddrInfo entries.
func (bc *BootstrapClient) BootstrapAddrs() []peer.AddrInfo {
	return bc.peers
}

// FetchPeers calls bootstrap.fetch on each configured bootstrap node and
// loads discovered peers into the cache.
func (bc *BootstrapClient) FetchPeers(ctx context.Context) {
	for _, bp := range bc.peers {
		if err := bc.h.Connect(ctx, bp); err != nil {
			continue
		}
		bc.fetchFrom(ctx, bp.ID)
	}
}

// RegisterSelf calls bootstrap.register to advertise this node to all configured bootstrap nodes.
func (bc *BootstrapClient) RegisterSelf(ctx context.Context) {
	for _, bp := range bc.peers {
		if err := bc.h.Connect(ctx, bp); err != nil {
			continue
		}
		bc.registerWith(ctx, bp.ID)
	}
}

func (bc *BootstrapClient) fetchFrom(ctx context.Context, id peer.ID) {
	tCtx, cancel := context.WithTimeout(ctx, bootstrapTimeout)
	defer cancel()

	s, err := bc.h.NewStream(tCtx, id, bootstrapProtocol)
	if err != nil {
		return
	}
	defer s.Close()

	req := bootstrapFetchRequest{Operation: "bootstrap.fetch", Limit: bootstrapMaxPeers}
	if err := json.NewEncoder(s).Encode(req); err != nil {
		return
	}
	_ = s.CloseWrite()

	var resp bootstrapFetchResponse
	if err := json.NewDecoder(s).Decode(&resp); err != nil {
		return
	}

	for _, r := range resp.Peers {
		if r.PeerID == bc.h.ID().String() {
			continue
		}
		rec := PeerRecord{
			PeerID:    r.PeerID,
			Addresses: r.Addrs,
			Role:      r.Role,
			Source:    "bootstrap",
		}
		bc.cache.Put(rec, addrTTL)
	}
}

func (bc *BootstrapClient) registerWith(ctx context.Context, id peer.ID) {
	tCtx, cancel := context.WithTimeout(ctx, bootstrapTimeout)
	defer cancel()

	s, err := bc.h.NewStream(tCtx, id, bootstrapProtocol)
	if err != nil {
		return
	}
	defer s.Close()

	addrs := bc.h.Addrs()
	addrStrs := make([]string, len(addrs))
	for i, a := range addrs {
		addrStrs[i] = a.String()
	}

	req := bootstrapRegisterRequest{
		Operation: "bootstrap.register",
		PeerID:    bc.h.ID().String(),
		Addrs:     addrStrs,
	}
	_ = json.NewEncoder(s).Encode(req)
	_ = s.CloseWrite()
}

// BootstrapServer handles incoming bootstrap protocol streams.
// It maintains a registry of peers and serves them on request.
type BootstrapServer struct {
	h     libp2phost.Host
	cache *Cache
}

// NewBootstrapServer creates a BootstrapServer and registers the protocol handler.
func NewBootstrapServer(h libp2phost.Host, cache *Cache) *BootstrapServer {
	bs := &BootstrapServer{h: h, cache: cache}
	h.SetStreamHandler(bootstrapProtocol, bs.handleStream)
	return bs
}

func (bs *BootstrapServer) handleStream(s network.Stream) {
	defer s.Close()
	_ = s.SetDeadline(time.Now().Add(bootstrapTimeout))

	// Decode the full message into a raw map so we can inspect Operation and
	// then unmarshal into the specific request type — all in one JSON pass.
	var raw json.RawMessage
	if err := json.NewDecoder(s).Decode(&raw); err != nil {
		return
	}

	var opOnly struct {
		Operation string `json:"operation"`
	}
	if err := json.Unmarshal(raw, &opOnly); err != nil {
		return
	}

	switch opOnly.Operation {
	case "bootstrap.register":
		var req bootstrapRegisterRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return
		}
		// Use addresses supplied by the registering peer if provided;
		// fall back to peerstore addresses only when none were sent.
		addrStrs := req.Addrs
		if len(addrStrs) == 0 {
			remotePeer := s.Conn().RemotePeer()
			psAddrs := bs.h.Peerstore().Addrs(remotePeer)
			addrStrs = make([]string, len(psAddrs))
			for i, a := range psAddrs {
				addrStrs[i] = a.String()
			}
		}
		peerID := req.PeerID
		if peerID == "" {
			peerID = s.Conn().RemotePeer().String()
		}
		rec := PeerRecord{
			PeerID:    peerID,
			Addresses: addrStrs,
			Source:    "bootstrap",
		}
		bs.cache.Put(rec, addrTTL)

	case "bootstrap.fetch":
		all := bs.cache.All()
		peers := make([]bootstrapPeerEntry, 0, len(all))
		for _, r := range all {
			if len(peers) >= bootstrapMaxPeers {
				break
			}
			peers = append(peers, bootstrapPeerEntry{
				PeerID: r.PeerID,
				Addrs:  r.Addresses,
				Role:   r.Role,
			})
		}
		resp := bootstrapFetchResponse{Peers: peers}
		_ = s.CloseRead()
		_ = json.NewEncoder(s).Encode(resp)
	}
}

type bootstrapFetchRequest struct {
	Operation string `json:"operation"`
	Limit     int    `json:"limit"`
}

type bootstrapRegisterRequest struct {
	Operation string   `json:"operation"`
	PeerID    string   `json:"peer_id"`
	Addrs     []string `json:"addrs"`
}

type bootstrapPeerEntry struct {
	PeerID string   `json:"peer_id"`
	Addrs  []string `json:"addrs"`
	Role   string   `json:"role,omitempty"`
}

type bootstrapFetchResponse struct {
	Peers []bootstrapPeerEntry `json:"peers"`
}
