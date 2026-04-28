package discovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	zeroconf "github.com/libp2p/zeroconf/v2"
	libp2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
)

const (
	mdnsService  = "_aether._udp"
	mdnsDomain   = "local."
	mdnsInterval = 60 * time.Second
)

// MDNSService announces and browses for peers on the local network using
// custom TXT records (spec 31 §2): peer_id, addrs (CSV), role, caps (CSV).
type MDNSService struct {
	h        libp2phost.Host
	role     string
	caps     []string
	cache    *Cache
	server   *zeroconf.Server
	cancelFn context.CancelFunc
}

// NewMDNSService creates an MDNSService but does not start it.
func NewMDNSService(h libp2phost.Host, role string, caps []string, cache *Cache) *MDNSService {
	return &MDNSService{h: h, role: role, caps: caps, cache: cache}
}

// Start announces this node and begins browsing for peers.
func (s *MDNSService) Start(ctx context.Context) error {
	addrs := s.h.Addrs()
	addrStrs := make([]string, len(addrs))
	for i, a := range addrs {
		addrStrs[i] = a.String()
	}

	txt := []string{
		"peer_id=" + s.h.ID().String(),
		"addrs=" + strings.Join(addrStrs, ","),
		"role=" + s.role,
		"caps=" + strings.Join(s.caps, ","),
	}

	// Port 0 — mDNS itself doesn't use a fixed port; this is metadata-only.
	srv, err := zeroconf.Register(
		s.h.ID().String(), // instance name = peer ID (unique per node)
		mdnsService,
		mdnsDomain,
		0,
		txt,
		nil, // all interfaces
	)
	if err != nil {
		return fmt.Errorf("mdns: register: %w", err)
	}
	s.server = srv

	browseCtx, cancel := context.WithCancel(ctx)
	s.cancelFn = cancel
	go s.browse(browseCtx)
	return nil
}

// Close stops the announcement and the browse loop.
func (s *MDNSService) Close() {
	if s.cancelFn != nil {
		s.cancelFn()
	}
	if s.server != nil {
		s.server.Shutdown()
	}
}

func (s *MDNSService) browse(ctx context.Context) {
	ticker := time.NewTicker(mdnsInterval)
	defer ticker.Stop()

	s.doBrowse(ctx)
	for {
		select {
		case <-ticker.C:
			s.doBrowse(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *MDNSService) doBrowse(ctx context.Context) {
	entries := make(chan *zeroconf.ServiceEntry, 16)
	browseCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	go func() {
		_ = zeroconf.Browse(browseCtx, mdnsService, mdnsDomain, entries)
	}()

	for {
		select {
		case e, ok := <-entries:
			if !ok {
				return
			}
			s.handleEntry(e)
		case <-browseCtx.Done():
			return
		}
	}
}

func (s *MDNSService) handleEntry(e *zeroconf.ServiceEntry) {
	var peerID, addrsCSV, role, capsCSV string
	for _, t := range e.Text {
		switch {
		case strings.HasPrefix(t, "peer_id="):
			peerID = strings.TrimPrefix(t, "peer_id=")
		case strings.HasPrefix(t, "addrs="):
			addrsCSV = strings.TrimPrefix(t, "addrs=")
		case strings.HasPrefix(t, "role="):
			role = strings.TrimPrefix(t, "role=")
		case strings.HasPrefix(t, "caps="):
			capsCSV = strings.TrimPrefix(t, "caps=")
		}
	}

	if peerID == "" || peerID == s.h.ID().String() {
		return
	}

	// Validate peer ID is parseable.
	if _, err := peer.Decode(peerID); err != nil {
		return
	}

	var addrs []string
	if addrsCSV != "" {
		for _, a := range strings.Split(addrsCSV, ",") {
			a = strings.TrimSpace(a)
			parsed, err := ma.NewMultiaddr(a)
			if err != nil {
				continue
			}
			// Spec 31 §2.3: advertised addresses MUST include /p2p/<peer_id> suffix.
			if !strings.Contains(parsed.String(), "/p2p/") {
				continue
			}
			addrs = append(addrs, a)
		}
	}

	var caps []string
	if capsCSV != "" {
		for _, c := range strings.Split(capsCSV, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				caps = append(caps, c)
			}
		}
	}

	r := PeerRecord{
		PeerID:    peerID,
		Addresses: addrs,
		Role:      role,
		Caps:      caps,
		Source:    "mdns",
	}
	s.cache.Put(r, peerTTL)
}
