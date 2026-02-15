package phase4

import (
	"context"
	"fmt"
	"time"
)

// TraversalStage represents each step in the ordered connectivity fallback chain.
type TraversalStage string

const (
	TraversalStageDirect    TraversalStage = "direct"
	TraversalStageAutoNAT   TraversalStage = "autonat"
	TraversalStageHolePunch TraversalStage = "hole-punch"
	TraversalStageRelay     TraversalStage = "relay"
)

var defaultTraversalOrder = []TraversalStage{
	TraversalStageDirect,
	TraversalStageAutoNAT,
	TraversalStageHolePunch,
	TraversalStageRelay,
}

// ConnectivityReasonCode encodes why a stage succeeded or failed so diagnostics can report it.
type ConnectivityReasonCode string

const (
	ReasonDirectSuccess          ConnectivityReasonCode = "direct-success"
	ReasonDirectUnavailable      ConnectivityReasonCode = "direct-unavailable"
	ReasonAutoNATSuccess         ConnectivityReasonCode = "autonat-success"
	ReasonAutoNATFailure         ConnectivityReasonCode = "autonat-failure"
	ReasonHolePunchSuccess       ConnectivityReasonCode = "hole-punch-success"
	ReasonHolePunchFailure       ConnectivityReasonCode = "hole-punch-failure"
	ReasonTraversalTimeout       ConnectivityReasonCode = "traversal-timeout"
	ReasonStageUnavailable       ConnectivityReasonCode = "stage-unavailable"
	ReasonRelayFallbackActive    ConnectivityReasonCode = "relay-fallback-active"
	ReasonRelayReservationFailed ConnectivityReasonCode = "relay-reservation-failed"
	ReasonRelayReleased          ConnectivityReasonCode = "relay-released"
	ReasonRecoveryTriggered      ConnectivityReasonCode = "recovery-triggered"
)

// TraversalEvent records lifecycle transitions and fallback attempts for diagnostics.
type TraversalEvent struct {
	Stage   TraversalStage
	Reason  ConnectivityReasonCode
	Message string
	Time    time.Time
}

// TraversalReport exposes the state machine outcome for clients.
type TraversalReport struct {
	Stage  TraversalStage
	Reason ConnectivityReasonCode
	Events []TraversalEvent
	Relay  RelayReservationStatus
}

// RelayReservationState enumerates the lifecycle a relay reservation can emit.
type RelayReservationState string

const (
	RelayReservationStatePending  RelayReservationState = "pending"
	RelayReservationStateReserved RelayReservationState = "reserved"
	RelayReservationStateActive   RelayReservationState = "active"
	RelayReservationStateReleased RelayReservationState = "released"
	RelayReservationStateFailed   RelayReservationState = "failed"
)

// RelayReservationEvent captures lifecycle signals emitted while handling relay fallback.
type RelayReservationEvent struct {
	State   RelayReservationState
	Reason  ConnectivityReasonCode
	Message string
	Time    time.Time
}

// RelayReservationStatus exposes the current reservation information and history.
type RelayReservationStatus struct {
	State  RelayReservationState
	Active bool
	Events []RelayReservationEvent
}

// StageResult is the observable result of a traversal stage attempt.
type StageResult struct {
	Success bool
	Reason  ConnectivityReasonCode
	Message string
}

// StageAction lets the caller inject deterministic behavior for each stage.
type StageAction func(ctx context.Context) StageResult

// RelayReservationOutcome is the lifecycle emitted by a relay fallback attempt.
type RelayReservationOutcome struct {
	State  RelayReservationState
	Events []RelayReservationEvent
}

// RelayReservationAction lets callers simulate reservation lifecycle events.
type RelayReservationAction func(ctx context.Context) RelayReservationOutcome

// TraversalHooks allows tests to inject stage and relay behaviors without networking.
type TraversalHooks struct {
	StageActions map[TraversalStage]StageAction
	RelayAction  RelayReservationAction
}

// FallbackTimeoutPolicy controls how long each traversal stage has to complete.
type FallbackTimeoutPolicy struct {
	StageTimeouts map[TraversalStage]time.Duration
	RelayTimeout  time.Duration
}

// TimeoutFor returns a timeout for the requested stage, falling back to defaults when needed.
func (p FallbackTimeoutPolicy) TimeoutFor(stage TraversalStage) time.Duration {
	if stage == TraversalStageRelay && p.RelayTimeout > 0 {
		return p.RelayTimeout
	}
	if p.StageTimeouts != nil {
		if duration, ok := p.StageTimeouts[stage]; ok {
			return duration
		}
	}
	switch stage {
	case TraversalStageDirect:
		return 5 * time.Second
	case TraversalStageAutoNAT:
		return 5 * time.Second
	case TraversalStageHolePunch:
		return 6 * time.Second
	case TraversalStageRelay:
		return 10 * time.Second
	default:
		return 5 * time.Second
	}
}

// TraversalRunner orchestrates the ordered fallback and records diagnostics.
type TraversalRunner struct {
	policy FallbackTimeoutPolicy
	hooks  TraversalHooks
}

// NewTraversalRunner constructs a runner with policy defaults when omitted.
func NewTraversalRunner(policy FallbackTimeoutPolicy, hooks TraversalHooks) *TraversalRunner {
	return &TraversalRunner{policy: policy, hooks: hooks}
}

// Run executes the traversal state machine, returning diagnostics suitable for clients.
func (r *TraversalRunner) Run(ctx context.Context) TraversalReport {
	events := make([]TraversalEvent, 0, len(defaultTraversalOrder)+1)
	var relayStatus RelayReservationStatus
	var finalStage TraversalStage
	var finalReason ConnectivityReasonCode
	failedAttempts := 0
	for _, stage := range defaultTraversalOrder {
		if stage == TraversalStageRelay {
			finalStage = stage
			finalReason = r.appendRelayFallbackEvent(ctx, &events, &relayStatus)
			break
		}
		result := r.executeStage(ctx, stage)
		events = append(events, r.eventFromResult(stage, result))
		if result.Success {
			if failedAttempts > 0 {
				events = append(events, TraversalEvent{
					Stage:   stage,
					Reason:  ReasonRecoveryTriggered,
					Message: fmt.Sprintf("recovered connectivity via %s", stage),
					Time:    time.Now(),
				})
			}
			finalStage = stage
			finalReason = result.Reason
			return TraversalReport{
				Stage:  finalStage,
				Reason: finalReason,
				Events: events,
				Relay:  relayStatus,
			}
		}
		failedAttempts++
	}
	return TraversalReport{
		Stage:  finalStage,
		Reason: finalReason,
		Events: events,
		Relay:  relayStatus,
	}
}

func (r *TraversalRunner) executeStage(parent context.Context, stage TraversalStage) StageResult {
	action := r.hooks.stageAction(stage)
	stageCtx, cancel := r.stageContext(parent, stage)
	defer cancel()
	result := action(stageCtx)
	if result.Reason == "" {
		result.Reason = defaultReasonFor(stage, result.Success)
	}
	if !result.Success {
		if err := stageCtx.Err(); err != nil && result.Reason == ReasonTraversalTimeout {
			result.Message = fmt.Sprintf("%s", err)
		}
	}
	return result
}

func (r *TraversalRunner) eventFromResult(stage TraversalStage, result StageResult) TraversalEvent {
	message := result.Message
	if message == "" {
		message = fmt.Sprintf("stage %s reason %s", stage, result.Reason)
	}
	return TraversalEvent{
		Stage:   stage,
		Reason:  result.Reason,
		Message: message,
		Time:    time.Now(),
	}
}

func (r *TraversalRunner) appendRelayFallbackEvent(parent context.Context, events *[]TraversalEvent, relay *RelayReservationStatus) ConnectivityReasonCode {
	action := r.hooks.relayAction()
	stageCtx, cancel := r.stageContext(parent, TraversalStageRelay)
	defer cancel()
	outcome := action(stageCtx)
	for i := range outcome.Events {
		if outcome.Events[i].Time.IsZero() {
			outcome.Events[i].Time = time.Now()
		}
	}
	*relay = RelayReservationStatus{
		State:  outcome.State,
		Active: outcome.State == RelayReservationStateReserved || outcome.State == RelayReservationStateActive,
		Events: append([]RelayReservationEvent(nil), outcome.Events...),
	}
	reason := ReasonRelayFallbackActive
	if relay.State == RelayReservationStateFailed {
		reason = ReasonRelayReservationFailed
	}
	*events = append(*events, TraversalEvent{
		Stage:   TraversalStageRelay,
		Reason:  reason,
		Message: fmt.Sprintf("relay fallback finished in state %s", relay.State),
		Time:    time.Now(),
	})
	return reason
}

func (r *TraversalRunner) stageContext(parent context.Context, stage TraversalStage) (context.Context, context.CancelFunc) {
	duration := r.policy.TimeoutFor(stage)
	if duration <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, duration)
}

func (hooks TraversalHooks) stageAction(stage TraversalStage) StageAction {
	if hooks.StageActions != nil {
		if action, ok := hooks.StageActions[stage]; ok && action != nil {
			return action
		}
	}
	return func(ctx context.Context) StageResult {
		return StageResult{
			Success: false,
			Reason:  ReasonStageUnavailable,
			Message: fmt.Sprintf("no action configured for %s", stage),
		}
	}
}

func (hooks TraversalHooks) relayAction() RelayReservationAction {
	if hooks.RelayAction != nil {
		return hooks.RelayAction
	}
	return func(ctx context.Context) RelayReservationOutcome {
		return RelayReservationOutcome{
			State: RelayReservationStateFailed,
			Events: []RelayReservationEvent{{
				State:   RelayReservationStateFailed,
				Reason:  ReasonRelayReservationFailed,
				Time:    time.Now(),
				Message: "relay fallback not configured",
			}},
		}
	}
}

func defaultReasonFor(stage TraversalStage, success bool) ConnectivityReasonCode {
	if success {
		switch stage {
		case TraversalStageDirect:
			return ReasonDirectSuccess
		case TraversalStageAutoNAT:
			return ReasonAutoNATSuccess
		case TraversalStageHolePunch:
			return ReasonHolePunchSuccess
		default:
			return ReasonStageUnavailable
		}
	}
	switch stage {
	case TraversalStageDirect:
		return ReasonDirectUnavailable
	case TraversalStageAutoNAT:
		return ReasonAutoNATFailure
	case TraversalStageHolePunch:
		return ReasonHolePunchFailure
	default:
		return ReasonTraversalTimeout
	}
}
