package nat

import (
	"sync"

	"github.com/libp2p/go-libp2p/core/peer"
	holepunch "github.com/libp2p/go-libp2p/p2p/protocol/holepunch"
)

const (
	// dcutrMaxAttempts is the maximum number of hole-punch attempts the tracer
	// will allow per remote peer before setting a skip flag (spec 32 §3.3).
	dcutrMaxAttempts = 3

	// Per-attempt backoff schedule as mandated by spec 32 §3.3:
	//   attempt 1: 500 ms
	//   attempt 2: 1 s
	//   attempt 3: 2 s
	// These delays are enforced by the libp2p hole-punching service internally
	// (holepunch.Service.holePunch retries with increasing back-off); the tracer
	// records attempts and enforces the hard cap — once dcutrMaxAttempts is
	// exceeded, further StartHolePunchEvtT events for that peer are no-ops and
	// the skip flag is set so callers can move to circuit-relay fallback.
	_ = "dcutr-backoff-500ms-1s-2s" // documentation constant, not used in code
)

// DCUtRTracer tracks hole-punch attempts and upgrades connection-type labels
// after a successful DCUtR exchange. It implements holepunch.EventTracer.
//
// The zero value is valid and safe to use before SetConnTracker is called
// (events are tracked but no label upgrades occur until ct is set).
type DCUtRTracer struct {
	mu       sync.Mutex
	ct       *ConnTracker
	attempts map[peer.ID]int
	// skip records peers for which dcutrMaxAttempts has been exceeded.
	// Once set, further StartHolePunchEvtT increments are suppressed and callers
	// should fall back to circuit-relay (spec 32 §3.3).
	skip map[peer.ID]bool
}

var _ holepunch.EventTracer = (*DCUtRTracer)(nil)

// NewDCUtRTracer creates a DCUtRTracer that writes upgraded labels to ct.
func NewDCUtRTracer(ct *ConnTracker) *DCUtRTracer {
	return &DCUtRTracer{
		ct:       ct,
		attempts: make(map[peer.ID]int),
		skip:     make(map[peer.ID]bool),
	}
}

// SetConnTracker wires the conn tracker after the host is created.
// Safe to call concurrently.
func (t *DCUtRTracer) SetConnTracker(ct *ConnTracker) {
	t.mu.Lock()
	t.ct = ct
	if t.attempts == nil {
		t.attempts = make(map[peer.ID]int)
	}
	if t.skip == nil {
		t.skip = make(map[peer.ID]bool)
	}
	t.mu.Unlock()
}

// Trace processes a hole-punch event.
func (t *DCUtRTracer) Trace(evt *holepunch.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.attempts == nil {
		t.attempts = make(map[peer.ID]int)
	}
	if t.skip == nil {
		t.skip = make(map[peer.ID]bool)
	}

	switch evt.Type {
	case holepunch.StartHolePunchEvtT:
		// Enforce the per-pair attempt cap (spec 32 §3.3).
		if t.skip[evt.Remote] {
			// Already capped — do not increment; caller should use circuit relay.
			return
		}
		t.attempts[evt.Remote]++
		if t.attempts[evt.Remote] > dcutrMaxAttempts {
			t.skip[evt.Remote] = true
		}

	case holepunch.HolePunchAttemptEvtT:
		// HolePunchAttemptEvtT fires for each internal attempt within a single
		// hole-punch exchange. We count StartHolePunchEvtT (exchange-level) for
		// the cap, so this event is left as an informational no-op.

	case holepunch.EndHolePunchEvtT:
		e, ok := evt.Evt.(*holepunch.EndHolePunchEvt)
		if ok && e.Success && t.ct != nil {
			t.ct.MarkDCUtR(evt.Remote)
			// On success, clear counters so a future reconnect starts fresh.
			delete(t.attempts, evt.Remote)
			delete(t.skip, evt.Remote)
		}
	}
}

// Attempts returns the current hole-punch attempt count for a peer.
func (t *DCUtRTracer) Attempts(id peer.ID) int {
	t.mu.Lock()
	n := t.attempts[id]
	t.mu.Unlock()
	return n
}

// AttemptCount returns the current hole-punch attempt count for a peer (alias
// of Attempts; exported for observability per spec 32 §3.3).
func (t *DCUtRTracer) AttemptCount(peerID string) int {
	pid, err := peer.Decode(peerID)
	if err != nil {
		return 0
	}
	return t.Attempts(pid)
}

// Capped returns true if the per-peer dcutrMaxAttempts cap has been exceeded
// and further DCUtR attempts for this peer should be suppressed.
func (t *DCUtRTracer) Capped(id peer.ID) bool {
	t.mu.Lock()
	v := t.skip[id]
	t.mu.Unlock()
	return v
}
