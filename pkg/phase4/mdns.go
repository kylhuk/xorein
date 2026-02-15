package phase4

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

// MDNSConfig encapsulates the deterministic knobs for the LAN discovery module.
type MDNSConfig struct {
	Service      string        // optional service identifier for consumption in later phases
	Domain       string        // optional domain name for lookups
	StartupDelay time.Duration // how long to wait before considering discovery ready
	Disabled     bool          // set true in restricted environments to opt out
}

// TimeFunc allows injecting deterministic time sources for instrumentation.
type TimeFunc func() time.Time

// MDNSDiscovery manages LAN peer detection with timing, dedup, and instrumentation logs.
type MDNSDiscovery struct {
	cfg         MDNSConfig
	timeFn      TimeFunc
	logs        []string
	discovered  map[string]struct{}
	startTime   time.Time
	readyTime   time.Time
	started     bool
	disabled    bool
	readyLogged bool
}

// NewMDNSDiscovery creates a new deterministic discovery module.
func NewMDNSDiscovery(cfg MDNSConfig, timeFn TimeFunc) *MDNSDiscovery {
	if timeFn == nil {
		timeFn = time.Now
	}
	return &MDNSDiscovery{
		cfg:        cfg,
		timeFn:     timeFn,
		discovered: make(map[string]struct{}),
	}
}

// Start records the initial startup window and logs the startup state.
func (d *MDNSDiscovery) Start(ctx context.Context) error {
	if d.started {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	now := d.timeFn()
	delay := d.cfg.StartupDelay
	if delay < 0 {
		delay = 0
	}
	d.startTime = now
	d.readyTime = now.Add(delay)
	d.started = true
	if d.cfg.Disabled {
		d.disabled = true
		d.log(now, "mdns discovery disabled for restricted environment")
		return nil
	}
	d.log(now, fmt.Sprintf("mdns discovery startup started (delay=%s)", delay))
	return nil
}

// StartupReady reports whether the configured startup delay has elapsed and logs readiness once.
func (d *MDNSDiscovery) StartupReady() bool {
	if !d.started || d.disabled || d.readyLogged {
		return d.readyLogged
	}
	now := d.timeFn()
	if now.Before(d.readyTime) {
		return false
	}
	d.readyLogged = true
	d.log(now, "mdns discovery startup ready")
	return true
}

// AddPeer records a discovered peer if it is new.
func (d *MDNSDiscovery) AddPeer(peer string) bool {
	if !d.started || d.disabled {
		return false
	}
	normalized := strings.TrimSpace(peer)
	if normalized == "" {
		return false
	}
	if _, ok := d.discovered[normalized]; ok {
		return false
	}
	d.discovered[normalized] = struct{}{}
	d.log(d.timeFn(), fmt.Sprintf("mdns discovered peer %s (total=%d)", normalized, len(d.discovered)))
	return true
}

// KnownPeers returns a deterministic list of discovered peers.
func (d *MDNSDiscovery) KnownPeers() []string {
	peers := make([]string, 0, len(d.discovered))
	for peer := range d.discovered {
		peers = append(peers, peer)
	}
	sort.Strings(peers)
	return peers
}

// Logs returns an immutable copy of all instrumentation logs.
func (d *MDNSDiscovery) Logs() []string {
	copied := make([]string, len(d.logs))
	copy(copied, d.logs)
	return copied
}

// IsDisabled indicates whether discovery was suppressed.
func (d *MDNSDiscovery) IsDisabled() bool {
	return d.disabled
}

func (d *MDNSDiscovery) log(ts time.Time, message string) {
	d.logs = append(d.logs, fmt.Sprintf("%s %s", ts.Format(time.RFC3339Nano), message))
}
