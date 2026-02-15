package phase4

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

// DHTConfig controls the minimal private namespace behavior for P4-T2.
// The enum-style Namespace and bootstrap peers are kept immutable so every
// node observes the same deterministic fallback path.
type DHTConfig struct {
	Namespace       string
	BootstrapPeers  []string
	WarmupTarget    int
	PeerCachePath   string
	CacheRetention  time.Duration
	CacheMaxEntries int
}

// ProbeFunc lets callers simulate dialing bootstrap peers for unit testing or
// future networking wiring.
type ProbeFunc func(context.Context, string) bool

// BootstrapStatus describes the result of a single bootstrap attempt.
type BootstrapStatus struct {
	Peer      string
	Namespace string
	Success   bool
	Reason    string
}

// DHTService encapsulates the in-memory state for the private namespace DHT
// setup. It keeps a deterministic fallback, tracks warmup progress, and stores
// the last bootstrap status list.
type DHTService struct {
	cfg          DHTConfig
	probe        ProbeFunc
	normalized   []string
	fallback     []string
	cache        *LocalPeerCache
	discovered   []string
	routingCount int
	lastStatuses []BootstrapStatus
}

// NewDHTService instantiates the minimal service for P4-T2. It enforces namespace
// presence, normalizes the bootstrap list (ingestion path), and seeds the private
// fallback namespace derived solely from the namespace string.
func NewDHTService(cfg DHTConfig, probe ProbeFunc) (*DHTService, error) {
	if strings.TrimSpace(cfg.Namespace) == "" {
		return nil, fmt.Errorf("phase4: namespace is required")
	}
	if cfg.WarmupTarget < 0 {
		cfg.WarmupTarget = 0
	}
	normalized := normalizePeers(cfg.BootstrapPeers)
	svc := &DHTService{
		cfg:        cfg,
		probe:      probe,
		normalized: normalized,
		fallback:   deriveFallbackPeers(cfg.Namespace),
	}
	if strings.TrimSpace(cfg.PeerCachePath) != "" {
		svc.cache = NewLocalPeerCache(cfg.PeerCachePath, cfg.CacheRetention, cfg.CacheMaxEntries)
		cachedPeers, err := svc.cache.Load()
		if err != nil {
			return nil, fmt.Errorf("phase4: load peer cache: %w", err)
		}
		svc.normalized = preferCachedPeers(cachedPeers, normalized)
	}
	return svc, nil
}

// Bootstrap attempts to join the configured namespace using the normalized peers.
// If every configured peer fails, the deterministic fallback sequence is appended
// so node operators can always reason about why discovery failed and rely on a
// cached list. Failure handling is explicit through the returned status list.
func (s *DHTService) Bootstrap(ctx context.Context) []BootstrapStatus {
	statuses, joined := s.probeList(ctx, s.normalized, false)
	if len(statuses) == 0 {
		statuses, _ = s.probeList(ctx, s.fallback, true)
		s.lastStatuses = statuses
		return statuses
	}
	if !joined {
		fallbackStatuses, _ := s.probeList(ctx, s.fallback, true)
		statuses = append(statuses, fallbackStatuses...)
	}
	s.lastStatuses = statuses
	return statuses
}

// Warmup increments the routing table count by simulating contacts with discovered
// peers. Warmup stops once the configured target is reached or no peers are known.
func (s *DHTService) Warmup(ctx context.Context) int {
	target := s.cfg.WarmupTarget
	for s.routingCount < target && len(s.discovered) > 0 {
		select {
		case <-ctx.Done():
			return s.routingCount
		default:
		}
		s.routingCount++
	}
	return s.routingCount
}

// RoutingCount exposes the number of nodes contacted during warmup.
func (s *DHTService) RoutingCount() int {
	return s.routingCount
}

// KnownPeers returns the list of peers that responded positively during bootstrap.
func (s *DHTService) KnownPeers() []string {
	copied := make([]string, len(s.discovered))
	copy(copied, s.discovered)
	return copied
}

// LastBootstrapStatuses reports the last bootstrap attempt results.
func (s *DHTService) LastBootstrapStatuses() []BootstrapStatus {
	copied := make([]BootstrapStatus, len(s.lastStatuses))
	copy(copied, s.lastStatuses)
	return copied
}

func (s *DHTService) probeList(ctx context.Context, peers []string, fallback bool) ([]BootstrapStatus, bool) {
	statuses := make([]BootstrapStatus, 0, len(peers))
	joined := false
	for _, peer := range peers {
		select {
		case <-ctx.Done():
			statuses = append(statuses, BootstrapStatus{
				Peer:      peer,
				Namespace: s.cfg.Namespace,
				Success:   false,
				Reason:    "context cancelled",
			})
			return statuses, joined
		default:
		}

		reason := "probe not configured"
		success := false
		if s.probe != nil {
			success = s.probe(ctx, peer)
			if success {
				reason = "joined namespace"
			} else {
				reason = "peer unreachable"
			}
		}

		if fallback && !success && len(peers) > 0 {
			reason = "deterministic fallback peer unreachable"
		}

		statuses = append(statuses, BootstrapStatus{
			Peer:      peer,
			Namespace: s.cfg.Namespace,
			Success:   success,
			Reason:    reason,
		})

		if success {
			joined = true
			s.discovered = appendUnique(s.discovered, peer)
			s.recordPeerSuccess(peer)
		}
	}
	return statuses, joined
}

func (s *DHTService) recordPeerSuccess(peer string) {
	if s.cache == nil {
		return
	}
	_ = s.cache.RecordSuccess(peer)
}

func normalizePeers(peers []string) []string {
	seen := make(map[string]struct{}, len(peers))
	normalized := make([]string, 0, len(peers))
	for _, raw := range peers {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	sort.Strings(normalized)
	return normalized
}

func deriveFallbackPeers(namespace string) []string {
	hashed := sha256.Sum256([]byte(namespace))
	prefix := hex.EncodeToString(hashed[:4])
	peers := make([]string, 3)
	for i := range peers {
		peers[i] = fmt.Sprintf("%s-fallback-%s-%02d", namespace, prefix, i+1)
	}
	return peers
}

func appendUnique(list []string, peer string) []string {
	for _, existing := range list {
		if existing == peer {
			return list
		}
	}
	return append(list, peer)
}
