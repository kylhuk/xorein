package presence

import (
	"strings"
	"time"
)

type PublishReason string

const (
	PublishReasonInitial          PublishReason = "publish-initial"
	PublishReasonStateChanged     PublishReason = "publish-state-changed"
	PublishReasonStatusChanged    PublishReason = "publish-status-changed"
	PublishReasonHeartbeat        PublishReason = "publish-heartbeat"
	PublishReasonDebounced        PublishReason = "publish-debounced"
	PublishReasonThrottled        PublishReason = "publish-throttled"
	PublishReasonUnchanged        PublishReason = "publish-unchanged"
	PublishReasonInvalidTimestamp PublishReason = "publish-invalid-timestamp"
)

type CadencePolicy struct {
	DebounceWindow     time.Duration
	MinPublishInterval time.Duration
	HeartbeatInterval  time.Duration
}

type PublishTracker struct {
	LastPublishedAt time.Time
	LastState       State
	LastStatus      string
}

type PublishInput struct {
	ObservedAt time.Time
	State      State
	Status     string
}

type PublishDecision struct {
	Publish bool
	Reason  PublishReason
	Next    PublishTracker
}

func EvaluatePublishCadence(policy CadencePolicy, tracker PublishTracker, input PublishInput) PublishDecision {
	next := tracker
	normalizedStatus := strings.TrimSpace(input.Status)
	if input.ObservedAt.IsZero() {
		return PublishDecision{Publish: false, Reason: PublishReasonInvalidTimestamp, Next: next}
	}
	if tracker.LastPublishedAt.IsZero() {
		next.LastPublishedAt = input.ObservedAt
		next.LastState = input.State
		next.LastStatus = normalizedStatus
		return PublishDecision{Publish: true, Reason: PublishReasonInitial, Next: next}
	}

	elapsed := input.ObservedAt.Sub(tracker.LastPublishedAt)
	stateChanged := input.State != tracker.LastState
	statusChanged := normalizedStatus != tracker.LastStatus

	if stateChanged || statusChanged {
		if policy.DebounceWindow > 0 && elapsed < policy.DebounceWindow {
			return PublishDecision{Publish: false, Reason: PublishReasonDebounced, Next: next}
		}
		if policy.MinPublishInterval > 0 && elapsed < policy.MinPublishInterval {
			return PublishDecision{Publish: false, Reason: PublishReasonThrottled, Next: next}
		}
		next.LastPublishedAt = input.ObservedAt
		next.LastState = input.State
		next.LastStatus = normalizedStatus
		reason := PublishReasonStateChanged
		if !stateChanged && statusChanged {
			reason = PublishReasonStatusChanged
		}
		return PublishDecision{Publish: true, Reason: reason, Next: next}
	}

	if policy.HeartbeatInterval > 0 && elapsed >= policy.HeartbeatInterval {
		next.LastPublishedAt = input.ObservedAt
		return PublishDecision{Publish: true, Reason: PublishReasonHeartbeat, Next: next}
	}
	return PublishDecision{Publish: false, Reason: PublishReasonUnchanged, Next: next}
}

type StatusReason string

const (
	StatusReasonApplied      StatusReason = "status-applied"
	StatusReasonNoop         StatusReason = "status-noop"
	StatusReasonStaleVersion StatusReason = "status-stale-version"
	StatusReasonRedacted     StatusReason = "status-redacted"
	StatusReasonInvalid      StatusReason = "status-invalid"
)

type StatusRecord struct {
	Value     string
	Version   uint64
	UpdatedAt time.Time
	Redacted  bool
}

type StatusDecision struct {
	Applied     bool
	Invalidated bool
	Reason      StatusReason
	Next        StatusRecord
}

func ApplyStatusUpdate(local StatusRecord, incoming StatusRecord) StatusDecision {
	if incoming.Version == 0 || incoming.UpdatedAt.IsZero() {
		return StatusDecision{Applied: false, Invalidated: false, Reason: StatusReasonInvalid, Next: local}
	}
	if incoming.Version < local.Version {
		return StatusDecision{Applied: false, Invalidated: false, Reason: StatusReasonStaleVersion, Next: local}
	}
	if incoming.Version == local.Version {
		if !incoming.UpdatedAt.After(local.UpdatedAt) {
			return StatusDecision{Applied: false, Invalidated: false, Reason: StatusReasonNoop, Next: local}
		}
	}

	next := incoming
	next.Value = strings.TrimSpace(incoming.Value)
	if incoming.Redacted {
		next.Value = ""
		return StatusDecision{Applied: true, Invalidated: true, Reason: StatusReasonRedacted, Next: next}
	}
	if next.Value == local.Value && next.Version == local.Version && next.Redacted == local.Redacted {
		return StatusDecision{Applied: false, Invalidated: false, Reason: StatusReasonNoop, Next: local}
	}
	return StatusDecision{Applied: true, Invalidated: true, Reason: StatusReasonApplied, Next: next}
}

type PropagationReason string

const (
	PropagationReasonInitial          PropagationReason = "propagation-initial"
	PropagationReasonChanged          PropagationReason = "propagation-changed"
	PropagationReasonHeartbeat        PropagationReason = "propagation-heartbeat"
	PropagationReasonThrottled        PropagationReason = "propagation-throttled"
	PropagationReasonUnchanged        PropagationReason = "propagation-unchanged"
	PropagationReasonInvalidTimestamp PropagationReason = "propagation-invalid-timestamp"
)

type PropagationPolicy struct {
	MinInterval time.Duration
	MaxLatency  time.Duration
}

type PropagationDecision struct {
	Propagate bool
	Reason    PropagationReason
	NextSent  time.Time
}

func EvaluateStatusPropagation(policy PropagationPolicy, lastSent, now time.Time, changed bool) PropagationDecision {
	if now.IsZero() {
		return PropagationDecision{Propagate: false, Reason: PropagationReasonInvalidTimestamp, NextSent: lastSent}
	}
	if lastSent.IsZero() {
		return PropagationDecision{Propagate: true, Reason: PropagationReasonInitial, NextSent: now}
	}
	elapsed := now.Sub(lastSent)
	if changed {
		if policy.MinInterval > 0 && elapsed < policy.MinInterval {
			return PropagationDecision{Propagate: false, Reason: PropagationReasonThrottled, NextSent: lastSent}
		}
		return PropagationDecision{Propagate: true, Reason: PropagationReasonChanged, NextSent: now}
	}
	if policy.MaxLatency > 0 && elapsed >= policy.MaxLatency {
		return PropagationDecision{Propagate: true, Reason: PropagationReasonHeartbeat, NextSent: now}
	}
	return PropagationDecision{Propagate: false, Reason: PropagationReasonUnchanged, NextSent: lastSent}
}
