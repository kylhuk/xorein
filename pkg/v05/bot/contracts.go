package bot

import "strings"

// LifecycleStage captures deterministic phases for bot-triggered operations.
type LifecycleStage string

const (
	StageReceived  LifecycleStage = "stage.received"
	StageValidated LifecycleStage = "stage.validated"
	StageExecuting LifecycleStage = "stage.executing"
	StageCompleted LifecycleStage = "stage.completed"
)

// LifecycleReason scopes why an event or command advanced through a stage.
type LifecycleReason string

const (
	ReasonEventEnqueued       LifecycleReason = "reason.event.enqueued"
	ReasonEventHandled        LifecycleReason = "reason.event.handled"
	ReasonCommandInvoked      LifecycleReason = "reason.command.invoked"
	ReasonCommandRejected     LifecycleReason = "reason.command.rejected"
	ReasonCommandDeferred     LifecycleReason = "reason.command.deferred"
	ReasonAuthExpired         LifecycleReason = "reason.auth.expired"
	ReasonAuthMismatch        LifecycleReason = "reason.auth.mismatch"
	ReasonSecurityModeBlocked LifecycleReason = "reason.security.mode_blocked"
)

// AuthOutcome summarizes deterministic authentication results.
type AuthOutcome string

const (
	AuthSuccess      AuthOutcome = "auth.success"
	AuthChallenge    AuthOutcome = "auth.challenge"
	AuthFailure      AuthOutcome = "auth.failure"
	AuthNotAttempted AuthOutcome = "auth.not_attempted"
)

// SecurityMode mirrors intended transport gating for bot commands.
type SecurityMode string

const (
	SecurityModeClear     SecurityMode = "security.clear"
	SecurityModeEncrypted SecurityMode = "security.encrypted"
	SecurityModeDeferred  SecurityMode = "security.deferred"
)

var DefaultSecurityMode = SecurityModeClear
var DefaultSecurityModeWhitelist = []SecurityMode{SecurityModeClear}

// AllowsMode determines whether the provided mode is admitted by default deciders.
func AllowsMode(mode SecurityMode) bool {
	for _, candidate := range DefaultSecurityModeWhitelist {
		if candidate == mode {
			return true
		}
	}
	return false
}

// TrustClass tags trust for command routing.
type TrustClass string

const (
	TrustClassUnknown TrustClass = "trust.unknown"
	TrustClassLow     TrustClass = "trust.low"
	TrustClassMedium  TrustClass = "trust.medium"
	TrustClassHigh    TrustClass = "trust.high"
)

// CommandLifecycle captures the deterministic metadata for lifecycle events.
type CommandLifecycle struct {
	Stage  LifecycleStage
	Reason LifecycleReason
	Auth   AuthOutcome
}

func (c CommandLifecycle) IsTerminal() bool {
	return c.Stage == StageCompleted
}

func (c CommandLifecycle) Trust() TrustClass {
	if c.Auth == AuthSuccess {
		return TrustClassHigh
	}
	if c.Auth == AuthChallenge {
		return TrustClassMedium
	}
	if c.Auth == AuthFailure {
		return TrustClassLow
	}
	return TrustClassUnknown
}

func (c CommandLifecycle) IsPlaintextOnly() bool {
	return AllowsMode(DefaultSecurityMode)
}

// NormalizeReason returns a deterministic, lowercased representation.
func NormalizeReason(reason LifecycleReason) string {
	return strings.ToLower(string(reason))
}
