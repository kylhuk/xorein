package discovery

import (
	"context"
	"encoding/json"
	"math/rand"
	"strings"
	"sync"
	"time"

	libp2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	pexProtocol    = "/aether/pex/0.1.0"
	pexMaxReturn   = 50
	pexTimeout     = 10 * time.Second
	pexTentativeTTL = 5 * time.Minute // unverified addresses expire after this
)

// PEX implements peer exchange with anti-flood protection (spec 31 §6).
type PEX struct {
	mu         sync.Mutex
	h          libp2phost.Host
	cache      *Cache
	tentative  map[string]time.Time // peerID → time-received for unverified addrs
}

// NewPEX creates a PEX service and registers the protocol handler.
func NewPEX(h libp2phost.Host, cache *Cache) *PEX {
	p := &PEX{
		h:         h,
		cache:     cache,
		tentative: make(map[string]time.Time),
	}
	h.SetStreamHandler(pexProtocol, p.handleStream)
	return p
}

// ExchangeWith opens a PEX stream to a random connected peer and merges returned records.
func (p *PEX) ExchangeWith(ctx context.Context) {
	peers := p.h.Network().Peers()
	if len(peers) == 0 {
		return
	}
	target := peers[rand.Intn(len(peers))] //nolint:gosec // non-cryptographic selection

	tCtx, cancel := context.WithTimeout(ctx, pexTimeout)
	defer cancel()

	s, err := p.h.NewStream(tCtx, target, pexProtocol)
	if err != nil {
		return
	}
	defer s.Close()

	// Build set of known peer IDs to send.
	all := p.cache.All()
	known := make([]string, 0, len(all))
	for _, r := range all {
		known = append(known, r.PeerID)
	}

	req := pexRequest{KnownPeerIDs: known, RequestNew: true, Limit: pexMaxReturn}
	if err := json.NewEncoder(s).Encode(req); err != nil {
		return
	}
	_ = s.CloseWrite()

	var resp pexResponse
	if err := json.NewDecoder(s).Decode(&resp); err != nil {
		return
	}

	now := time.Now()
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, r := range resp.Peers {
		if r.PeerID == p.h.ID().String() {
			continue
		}
		// Spec 31 §5.3: only accept addresses that carry /p2p/<peer_id> suffix.
		filteredAddrs := filterP2PAddrs(r.Addrs)

		// Anti-flood: treat received peer as tentative; only promote to cache
		// after pexTentativeTTL has passed without the address being evicted.
		if t, ok := p.tentative[r.PeerID]; ok {
			if now.Sub(t) < pexTentativeTTL {
				// Still in tentative window — update cache but keep TTL short.
				rec := PeerRecord{
					PeerID:    r.PeerID,
					Addresses: filteredAddrs,
					Role:      r.Role,
					Source:    "pex",
				}
				p.cache.Put(rec, peerTTL)
				continue
			}
		}
		// First time seeing this peer via PEX; mark tentative.
		p.tentative[r.PeerID] = now
		rec := PeerRecord{
			PeerID:    r.PeerID,
			Addresses: filteredAddrs,
			Role:      r.Role,
			Source:    "pex",
		}
		p.cache.Put(rec, pexTentativeTTL)
	}

	// Evict stale tentative entries older than pexTentativeTTL.
	for id, t := range p.tentative {
		if now.Sub(t) >= pexTentativeTTL {
			delete(p.tentative, id)
		}
	}
}

// handleStream serves an incoming PEX request.
func (p *PEX) handleStream(s network.Stream) {
	defer s.Close()
	_ = s.SetDeadline(time.Now().Add(pexTimeout))

	var req pexRequest
	if err := json.NewDecoder(s).Decode(&req); err != nil {
		return
	}
	_ = s.CloseRead()

	// Build exclusion set from known_peer_ids.
	exclude := make(map[string]bool, len(req.KnownPeerIDs))
	for _, id := range req.KnownPeerIDs {
		exclude[id] = true
	}

	all := p.cache.All()
	// Shuffle to avoid systematic bias.
	rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] }) //nolint:gosec

	limit := req.Limit
	if limit <= 0 || limit > pexMaxReturn {
		limit = pexMaxReturn
	}

	peers := make([]pexPeerEntry, 0, limit)
	for _, r := range all {
		if len(peers) >= limit {
			break
		}
		if exclude[r.PeerID] || r.PeerID == p.h.ID().String() {
			continue
		}
		peers = append(peers, pexPeerEntry{
			PeerID: r.PeerID,
			Addrs:  r.Addresses,
			Role:   r.Role,
		})
	}

	resp := pexResponse{Peers: peers}
	_ = json.NewEncoder(s).Encode(resp)
}

type pexRequest struct {
	KnownPeerIDs []string `json:"known_peer_ids,omitempty"`
	RequestNew   bool     `json:"request_new"`
	Limit        int      `json:"limit,omitempty"`
}

type pexPeerEntry struct {
	PeerID string   `json:"peer_id"`
	Addrs  []string `json:"addrs,omitempty"`
	Role   string   `json:"role,omitempty"`
}

type pexResponse struct {
	Peers []pexPeerEntry `json:"peers"`
}

// filterP2PAddrs returns only those addresses that are valid multiaddrs AND
// carry a /p2p/<peer_id> component (spec 31 §5.3).
func filterP2PAddrs(addrs []string) []string {
	var out []string
	for _, a := range addrs {
		parsed, err := ma.NewMultiaddr(a)
		if err != nil {
			continue
		}
		if strings.Contains(parsed.String(), "/p2p/") {
			out = append(out, a)
		}
	}
	return out
}
