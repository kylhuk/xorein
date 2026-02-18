package moderation

import (
	"fmt"
	"sort"
)

// EventType describes the kind of signed moderation update.
type EventType string

const (
	EventRedaction EventType = "redaction"
	EventTimeout   EventType = "timeout"
	EventBan       EventType = "ban"
	EventSlowMode  EventType = "slow_mode"
	EventLockdown  EventType = "lockdown"
)

// RejectionReason labels deterministic rejection outcomes for moderation events.
type RejectionReason string

const (
	ReasonAccepted         RejectionReason = "accepted"
	ReasonInvalidSignature RejectionReason = "invalid-signature"
	ReasonStaleEvent       RejectionReason = "stale-event"
	ReasonDuplicateEvent   RejectionReason = "duplicate-event"
	ReasonMissingRoom      RejectionReason = "missing-room"
	ReasonUnknownAction    RejectionReason = "unknown-action"
)

// SignedEvent is the canonical payload used by v17 moderation.
type SignedEvent struct {
	ID        string
	Room      string
	Actor     string
	Target    string
	Type      EventType
	Timestamp int64
	Signature string
}

// Result encodes the deterministic outcome for an event.
type Result struct {
	Accepted bool
	Reason   RejectionReason
	Trace    string
}

// Engine maintains deterministic ordering and deduplication state.
type Engine struct {
	seen          map[string]struct{}
	allowed       map[EventType]struct{}
	sequenceTrace []string
	lastTimestamp map[string]int64
}

// NewEngine initializes a new moderation engine with deterministic defaults.
func NewEngine() *Engine {
	return &Engine{
		seen:          make(map[string]struct{}),
		allowed:       map[EventType]struct{}{EventRedaction: {}, EventTimeout: {}, EventBan: {}, EventSlowMode: {}, EventLockdown: {}},
		lastTimestamp: make(map[string]int64),
	}
}

// Apply runs deterministic validation and ordering checks.
func (e *Engine) Apply(event SignedEvent) Result {
	if event.Room == "" {
		return Result{Accepted: false, Reason: ReasonMissingRoom, Trace: "room required for ordering"}
	}
	if !e.validateSignature(event) {
		return Result{Accepted: false, Reason: ReasonInvalidSignature, Trace: "signature mismatch"}
	}
	if _, seen := e.seen[event.ID]; seen {
		return Result{Accepted: false, Reason: ReasonDuplicateEvent, Trace: "event already applied"}
	}
	if ts := e.lastTimestamp[event.Room]; event.Timestamp <= ts {
		return Result{Accepted: false, Reason: ReasonStaleEvent, Trace: fmt.Sprintf("timestamp %d <= last %d", event.Timestamp, ts)}
	}
	if !e.isAllowedType(event.Type) {
		return Result{Accepted: false, Reason: ReasonUnknownAction, Trace: fmt.Sprintf("unsupported %s", event.Type)}
	}
	e.seen[event.ID] = struct{}{}
	e.lastTimestamp[event.Room] = event.Timestamp
	e.sequenceTrace = append(e.sequenceTrace, fmt.Sprintf("[%s] %s -> %s", event.Type, event.Actor, event.Target))
	return Result{Accepted: true, Reason: ReasonAccepted, Trace: fmt.Sprintf("applied %s", event.Type)}
}

// SeenCount reports how many unique events have been applied.
func (e *Engine) SeenCount() int {
	return len(e.seen)
}

// LastTimestamp reports the most recent timestamp observed for a room.
func (e *Engine) LastTimestamp(room string) int64 {
	return e.lastTimestamp[room]
}

func (e *Engine) validateSignature(event SignedEvent) bool {
	return event.Signature == expectedSignature(event.Actor)
}

func (e *Engine) isAllowedType(t EventType) bool {
	_, ok := e.allowed[t]
	return ok
}

func expectedSignature(actor string) string {
	return fmt.Sprintf("sig:%s", actor)
}

// SequenceTrace returns the order of applied events for diagnostics.
func (e *Engine) SequenceTrace() []string {
	copyTrace := append([]string(nil), e.sequenceTrace...)
	sort.Strings(copyTrace)
	return copyTrace
}
