package groupdm

import "sort"

const (
	MemberCap           = 50
	MemberWarnThreshold = 40
)

type LimitReason string

const (
	ReasonMemberCapExceeded LimitReason = "member-cap-exceeded"
	ReasonMemberCapWarning  LimitReason = "member-cap-warning"
	ReasonMemberCapOK       LimitReason = "member-cap-ok"
)

type MemberCapStatus struct {
	Allowed bool
	Warning bool
	Reason  LimitReason
}

func ValidateMemberCap(current, adding int) MemberCapStatus {
	next := current + adding
	if next > MemberCap {
		return MemberCapStatus{Allowed: false, Reason: ReasonMemberCapExceeded}
	}
	if next >= MemberWarnThreshold {
		return MemberCapStatus{Allowed: true, Warning: true, Reason: ReasonMemberCapWarning}
	}
	return MemberCapStatus{Allowed: true, Reason: ReasonMemberCapOK}
}

type MembershipEvent struct {
	EventID  string
	Sequence uint64
}

func NormalizeMembershipEvents(events []MembershipEvent) []MembershipEvent {
	out := append([]MembershipEvent(nil), events...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Sequence == out[j].Sequence {
			return out[i].EventID < out[j].EventID
		}
		return out[i].Sequence < out[j].Sequence
	})
	return out
}

type MemberState string

const (
	MemberStateNone    MemberState = "none"
	MemberStateInvited MemberState = "invited"
	MemberStateMember  MemberState = "member"
	MemberStateLeft    MemberState = "left"
	MemberStateRemoved MemberState = "removed"
)

type MembershipEventType string

const (
	EventInvite MembershipEventType = "invite"
	EventJoin   MembershipEventType = "join"
	EventLeave  MembershipEventType = "leave"
	EventRemove MembershipEventType = "remove"
)

type TransitionReason string

const (
	ReasonTransitionApplied TransitionReason = "transition-applied"
	ReasonInvalidTransition TransitionReason = "invalid-transition"
)

type TransitionResult struct {
	Next   MemberState
	Reason TransitionReason
}

func EvaluateMemberTransition(current MemberState, event MembershipEventType) TransitionResult {
	if allowed, next := transitionTarget(current, event); allowed {
		return TransitionResult{Next: next, Reason: ReasonTransitionApplied}
	}
	return TransitionResult{Next: current, Reason: ReasonInvalidTransition}
}

func transitionTarget(current MemberState, event MembershipEventType) (bool, MemberState) {
	switch current {
	case MemberStateNone:
		if event == EventInvite {
			return true, MemberStateInvited
		}
	case MemberStateInvited:
		if event == EventJoin {
			return true, MemberStateMember
		}
		if event == EventRemove {
			return true, MemberStateRemoved
		}
	case MemberStateMember:
		if event == EventLeave {
			return true, MemberStateLeft
		}
		if event == EventRemove {
			return true, MemberStateRemoved
		}
	case MemberStateLeft:
		if event == EventJoin {
			return true, MemberStateMember
		}
	}
	return false, current
}

type SenderKeyDistributionReason string

const (
	ReasonSenderKeyAccepted             SenderKeyDistributionReason = "sender-key-accepted"
	ReasonSenderKeyInvalidEnvelope      SenderKeyDistributionReason = "sender-key-invalid-envelope"
	ReasonSenderKeyUnauthorizedSender   SenderKeyDistributionReason = "sender-key-unauthorized-sender"
	ReasonSenderKeyMissingSignature     SenderKeyDistributionReason = "sender-key-missing-signature"
	ReasonSenderKeyReplayUnprotected    SenderKeyDistributionReason = "sender-key-replay-unprotected"
	ReasonSenderKeyRetryScheduled       SenderKeyDistributionReason = "sender-key-retry-scheduled"
	ReasonSenderKeyRetryExhausted       SenderKeyDistributionReason = "sender-key-retry-exhausted"
	ReasonSenderKeyRetryTerminalFailure SenderKeyDistributionReason = "sender-key-retry-terminal-failure"
)

type SenderKeyEnvelope struct {
	EnvelopeID          string
	GroupID             string
	SenderID            string
	RecipientID         string
	Epoch               uint64
	Nonce               string
	Ciphertext          []byte
	SignatureAlgorithm  string
	Signature           []byte
	ReplayProtectionTag string
}

type SenderKeyDistributionDecision struct {
	Accepted bool
	Retry    bool
	Reason   SenderKeyDistributionReason
}

func ValidateSenderKeyEnvelope(envelope SenderKeyEnvelope, authorizedSenders map[string]struct{}) SenderKeyDistributionDecision {
	if envelope.EnvelopeID == "" || envelope.GroupID == "" || envelope.SenderID == "" || envelope.RecipientID == "" || envelope.Epoch == 0 || envelope.Nonce == "" || len(envelope.Ciphertext) == 0 {
		return SenderKeyDistributionDecision{Accepted: false, Reason: ReasonSenderKeyInvalidEnvelope}
	}
	if _, ok := authorizedSenders[envelope.SenderID]; !ok {
		return SenderKeyDistributionDecision{Accepted: false, Reason: ReasonSenderKeyUnauthorizedSender}
	}
	if envelope.SignatureAlgorithm == "" || len(envelope.Signature) == 0 {
		return SenderKeyDistributionDecision{Accepted: false, Reason: ReasonSenderKeyMissingSignature}
	}
	if envelope.ReplayProtectionTag == "" {
		return SenderKeyDistributionDecision{Accepted: false, Reason: ReasonSenderKeyReplayUnprotected}
	}
	return SenderKeyDistributionDecision{Accepted: true, Reason: ReasonSenderKeyAccepted}
}

func EvaluateSenderKeyBootstrapRetry(attempt, maxAttempts uint32, recoverable bool) SenderKeyDistributionDecision {
	if !recoverable {
		return SenderKeyDistributionDecision{Accepted: false, Retry: false, Reason: ReasonSenderKeyRetryTerminalFailure}
	}
	if maxAttempts == 0 {
		maxAttempts = 1
	}
	if attempt+1 >= maxAttempts {
		return SenderKeyDistributionDecision{Accepted: false, Retry: false, Reason: ReasonSenderKeyRetryExhausted}
	}
	return SenderKeyDistributionDecision{Accepted: false, Retry: true, Reason: ReasonSenderKeyRetryScheduled}
}

type RekeyTrigger string

const (
	RekeyTriggerMemberAdded     RekeyTrigger = "member-added"
	RekeyTriggerMemberRemoved   RekeyTrigger = "member-removed"
	RekeyTriggerMemberLeft      RekeyTrigger = "member-left"
	RekeyTriggerCompromise      RekeyTrigger = "compromise-suspected"
	RekeyTriggerRejoinAfterDrop RekeyTrigger = "rejoin-after-removal"
)

type RekeyReason string

const (
	ReasonRekeyRequiredMembershipChange RekeyReason = "rekey-required-membership-change"
	ReasonRekeyRequiredCompromise       RekeyReason = "rekey-required-compromise"
	ReasonRekeyRequiredRejoinGate       RekeyReason = "rekey-required-rejoin-gate"
	ReasonRekeyNotRequired              RekeyReason = "rekey-not-required"
)

type RekeyDecision struct {
	Mandatory     bool
	RejoinAllowed bool
	Reason        RekeyReason
}

func EvaluateRekeyTrigger(trigger RekeyTrigger) RekeyDecision {
	switch trigger {
	case RekeyTriggerMemberAdded, RekeyTriggerMemberRemoved, RekeyTriggerMemberLeft:
		return RekeyDecision{Mandatory: true, RejoinAllowed: true, Reason: ReasonRekeyRequiredMembershipChange}
	case RekeyTriggerCompromise:
		return RekeyDecision{Mandatory: true, RejoinAllowed: false, Reason: ReasonRekeyRequiredCompromise}
	case RekeyTriggerRejoinAfterDrop:
		return RekeyDecision{Mandatory: true, RejoinAllowed: false, Reason: ReasonRekeyRequiredRejoinGate}
	default:
		return RekeyDecision{Mandatory: false, RejoinAllowed: true, Reason: ReasonRekeyNotRequired}
	}
}

func CanRejoinAfterRemoval(removedEpoch, currentEpoch uint64, rekeyCompleted bool) bool {
	if !rekeyCompleted {
		return false
	}
	return currentEpoch > removedEpoch
}

type GroupTransportPath string

const (
	GroupTransportPathDirect  GroupTransportPath = "direct"
	GroupTransportPathOffline GroupTransportPath = "offline"
	GroupTransportPathReject  GroupTransportPath = "reject"
)

type GroupTransportReason string

const (
	ReasonGroupTransportDirectAvailable GroupTransportReason = "group-transport-direct-available"
	ReasonGroupTransportOfflineFallback GroupTransportReason = "group-transport-offline-fallback"
	ReasonGroupTransportRejected        GroupTransportReason = "group-transport-rejected"
)

type GroupTransportDecision struct {
	Path   GroupTransportPath
	Reason GroupTransportReason
}

func SelectGroupTransport(peerOnline, directCapable, offlineCapable bool) GroupTransportDecision {
	if peerOnline && directCapable {
		return GroupTransportDecision{Path: GroupTransportPathDirect, Reason: ReasonGroupTransportDirectAvailable}
	}
	if offlineCapable {
		return GroupTransportDecision{Path: GroupTransportPathOffline, Reason: ReasonGroupTransportOfflineFallback}
	}
	return GroupTransportDecision{Path: GroupTransportPathReject, Reason: ReasonGroupTransportRejected}
}

type HistorySyncReason string

const (
	ReasonHistoryOK          HistorySyncReason = "history-ok"
	ReasonHistoryLocked      HistorySyncReason = "history-locked"
	ReasonHistoryWindowBound HistorySyncReason = "history-window-bounded"
	ReasonHistoryInvalid     HistorySyncReason = "history-invalid-range"
)

type HistorySyncDecision struct {
	FromSeq      uint64
	ToSeq        uint64
	Locked       bool
	Reason       HistorySyncReason
	FromJoinTime bool
}

func ResolveHistorySync(joinSeq, requestFrom, requestTo, latestSeq, maxWindow uint64, hasPriorEpochKeys bool) HistorySyncDecision {
	if latestSeq == 0 {
		return HistorySyncDecision{Reason: ReasonHistoryInvalid}
	}
	if requestTo == 0 || requestTo > latestSeq {
		requestTo = latestSeq
	}
	if requestFrom == 0 {
		requestFrom = joinSeq
	}
	if requestFrom > requestTo {
		return HistorySyncDecision{FromSeq: requestTo, ToSeq: requestTo, Reason: ReasonHistoryInvalid}
	}

	from := requestFrom
	locked := false
	reason := ReasonHistoryOK
	fromJoin := false

	if !hasPriorEpochKeys && from < joinSeq {
		from = joinSeq
		locked = true
		reason = ReasonHistoryLocked
		fromJoin = true
	}
	if maxWindow > 0 && requestTo >= from {
		window := requestTo - from + 1
		if window > maxWindow {
			boundedFrom := requestTo - maxWindow + 1
			if boundedFrom < from {
				boundedFrom = from
			}
			from = boundedFrom
			if reason == ReasonHistoryOK {
				reason = ReasonHistoryWindowBound
			}
		}
	}

	if from > requestTo {
		return HistorySyncDecision{FromSeq: requestTo, ToSeq: requestTo, Locked: locked, Reason: ReasonHistoryInvalid, FromJoinTime: fromJoin}
	}
	return HistorySyncDecision{FromSeq: from, ToSeq: requestTo, Locked: locked, Reason: reason, FromJoinTime: fromJoin}
}

const HistoryDisclosureCodeNotTransferable = "history-not-transferable"

type GrowthReason string

const (
	ReasonGrowthOK              GrowthReason = "growth-ok"
	ReasonGrowthWarning         GrowthReason = "growth-warning"
	ReasonGrowthConvertRequired GrowthReason = "growth-convert-required"
)

type GrowthDecision struct {
	Allowed             bool
	Warning             bool
	ConvertToServer     bool
	HistoryTransferable bool
	Reason              GrowthReason
	DisclosureCode      string
}

type ConvertPlan struct {
	CreateServer         bool
	CreateInitialChannel bool
	MigrateMemberList    bool
	PostSystemNotice     bool
	HistoryTransferable  bool
	DisclosureCode       string
}

func EvaluateGrowth(currentMembers, addingMembers int) GrowthDecision {
	status := ValidateMemberCap(currentMembers, addingMembers)
	if !status.Allowed {
		return GrowthDecision{
			Allowed:             false,
			ConvertToServer:     true,
			HistoryTransferable: false,
			Reason:              ReasonGrowthConvertRequired,
			DisclosureCode:      HistoryDisclosureCodeNotTransferable,
		}
	}
	if status.Warning {
		return GrowthDecision{Allowed: true, Warning: true, Reason: ReasonGrowthWarning}
	}
	return GrowthDecision{Allowed: true, Reason: ReasonGrowthOK}
}

func BuildConvertPlan() ConvertPlan {
	return ConvertPlan{
		CreateServer:         true,
		CreateInitialChannel: true,
		MigrateMemberList:    true,
		PostSystemNotice:     true,
		HistoryTransferable:  false,
		DisclosureCode:       HistoryDisclosureCodeNotTransferable,
	}
}
