package automod

import "github.com/aether/code_aether/pkg/v04/securitymode"

type TriggerType string
type ActionType string
type BypassReason string
type FailureReason string

const (
	TriggerRate       TriggerType = "rate"
	TriggerKeyword    TriggerType = "keyword"
	TriggerExtensible TriggerType = "extensible"

	ActionWarn       ActionType = "warn"
	ActionQuarantine ActionType = "quarantine"
	ActionBlock      ActionType = "block"
	ActionEscalate   ActionType = "escalate"

	BypassManualOverride BypassReason = "manual_override"
	BypassPolicyExempt   BypassReason = "policy_exempt"
	BypassEvidenceLag    BypassReason = "evidence_lag"

	FailurePolicyMissing FailureReason = "policy_missing"
	FailureEvidenceError FailureReason = "evidence_error"
	FailureModeMismatch  FailureReason = "mode_mismatch"
)

var triggerPriority = map[TriggerType]int{
	TriggerRate:       3,
	TriggerKeyword:    2,
	TriggerExtensible: 1,
}

func TriggerPrecedence(t TriggerType) int {
	if prec, ok := triggerPriority[t]; ok {
		return prec
	}
	return 0
}

func ActionFor(trigger TriggerType, mode securitymode.ChannelMode) ActionType {
	switch trigger {
	case TriggerRate:
		if mode == securitymode.ModeClear || mode == securitymode.ModeChannel {
			return ActionWarn
		}
		return ActionQuarantine
	case TriggerKeyword:
		if mode == securitymode.ModeE2EE {
			return ActionEscalate
		}
		return ActionBlock
	case TriggerExtensible:
		return ActionEscalate
	default:
		return ActionWarn
	}
}

func BypassPermitted(reason BypassReason, mode securitymode.ChannelMode) bool {
	if reason == BypassManualOverride {
		return mode == securitymode.ModeClear || mode == securitymode.ModeChannel
	}
	if reason == BypassPolicyExempt {
		return mode == securitymode.ModeCrowd || mode == securitymode.ModeTree
	}
	return reason == BypassEvidenceLag && mode != securitymode.ModeE2EE
}

func FailureMessage(reason FailureReason, mode securitymode.ChannelMode) string {
	switch reason {
	case FailurePolicyMissing:
		return "policy version unavailable"
	case FailureEvidenceError:
		return "audit evidence incomplete for " + string(mode)
	case FailureModeMismatch:
		return "non-compliant security mode for automation"
	default:
		return "unknown failure"
	}
}

func RequiresEvidence(trigger TriggerType, mode securitymode.ChannelMode) bool {
	if mode == securitymode.ModeE2EE {
		return trigger != TriggerRate
	}
	return trigger != TriggerExtensible
}
