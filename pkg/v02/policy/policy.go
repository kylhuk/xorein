package policy

import (
	"strings"
	"time"

	"github.com/aether/code_aether/pkg/v02/rbac"
)

type MentionType string

const (
	MentionTypeUser     MentionType = "user"
	MentionTypeRole     MentionType = "role"
	MentionTypeEveryone MentionType = "everyone"
	MentionTypeHere     MentionType = "here"
)

type MentionToken struct {
	Type MentionType
	Raw  string
}

func ParseMentions(text string) []MentionToken {
	parts := strings.Fields(text)
	out := make([]MentionToken, 0)
	for _, part := range parts {
		if strings.HasPrefix(part, `\\@`) {
			continue
		}
		if !strings.HasPrefix(part, "@") || len(part) < 2 {
			continue
		}
		tail := strings.Trim(strings.TrimPrefix(part, "@"), " ,.!?;:")
		typeValue := MentionTypeUser
		switch tail {
		case "here":
			typeValue = MentionTypeHere
		case "everyone":
			typeValue = MentionTypeEveryone
		case "role":
			typeValue = MentionTypeRole
		}
		out = append(out, MentionToken{Type: typeValue, Raw: part})
	}
	return out
}

type MentionTargetStatus string

const (
	MentionTargetValid   MentionTargetStatus = "valid"
	MentionTargetStale   MentionTargetStatus = "stale"
	MentionTargetUnknown MentionTargetStatus = "unknown"
)

type MentionDirectory struct {
	Users map[string]MentionTargetStatus
	Roles map[string]MentionTargetStatus
}

type MentionResolution struct {
	Type   MentionType
	Target string
	Status MentionTargetStatus
	Notify bool
	Render string
}

func ResolveMentionToken(token MentionToken, directory MentionDirectory) MentionResolution {
	target := normalizeMentionKey(token.Raw)
	raw := token.Raw
	if raw == "" {
		raw = "@" + target
	}

	switch token.Type {
	case MentionTypeEveryone, MentionTypeHere:
		if target == "" {
			target = string(token.Type)
		}
		return MentionResolution{
			Type:   token.Type,
			Target: target,
			Status: MentionTargetValid,
			Notify: true,
			Render: "@" + target,
		}
	case MentionTypeRole:
		status := lookupMentionStatus(directory.Roles, target)
		if status == MentionTargetUnknown {
			return fallbackMention(MentionTypeRole, target, raw)
		}
		return resolvedMention(MentionTypeRole, target, status, raw)
	default:
		userStatus := lookupMentionStatus(directory.Users, target)
		roleStatus := lookupMentionStatus(directory.Roles, target)

		switch {
		case userStatus == MentionTargetValid:
			return resolvedMention(MentionTypeUser, target, userStatus, raw)
		case roleStatus == MentionTargetValid:
			return resolvedMention(MentionTypeRole, target, roleStatus, raw)
		case userStatus == MentionTargetStale:
			return resolvedMention(MentionTypeUser, target, userStatus, raw)
		case roleStatus == MentionTargetStale:
			return resolvedMention(MentionTypeRole, target, roleStatus, raw)
		default:
			return fallbackMention(MentionTypeUser, target, raw)
		}
	}
}

func normalizeMentionKey(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "@") {
		return ""
	}
	return strings.Trim(strings.TrimPrefix(trimmed, "@"), " ,.!?;:")
}

func lookupMentionStatus(values map[string]MentionTargetStatus, key string) MentionTargetStatus {
	if key == "" || values == nil {
		return MentionTargetUnknown
	}
	status, ok := values[key]
	if !ok {
		return MentionTargetUnknown
	}
	if status == MentionTargetValid || status == MentionTargetStale {
		return status
	}
	return MentionTargetUnknown
}

func resolvedMention(mType MentionType, target string, status MentionTargetStatus, raw string) MentionResolution {
	render := "@" + target
	notify := status == MentionTargetValid
	if !notify {
		render = raw
	}
	return MentionResolution{
		Type:   mType,
		Target: target,
		Status: status,
		Notify: notify,
		Render: render,
	}
}

func fallbackMention(mType MentionType, target string, raw string) MentionResolution {
	return MentionResolution{
		Type:   mType,
		Target: target,
		Status: MentionTargetUnknown,
		Notify: false,
		Render: raw,
	}
}

type PolicyReason string

const (
	ReasonMentionAuthorized          PolicyReason = "mention-authorized"
	ReasonMentionUnauthorized        PolicyReason = "mention-unauthorized"
	ReasonSlowModeActive             PolicyReason = "slow-mode-active"
	ReasonSlowModeBypass             PolicyReason = "slow-mode-bypass"
	ReasonSlowModePass               PolicyReason = "slow-mode-pass"
	ReasonSlowModeReplayDuplicate    PolicyReason = "slow-mode-replay-duplicate"
	ReasonSlowModeReplayStaleEvent   PolicyReason = "slow-mode-replay-stale-event"
	ReasonModerationAllowed          PolicyReason = "moderation-allowed"
	ReasonModerationForbiddenTarget  PolicyReason = "moderation-forbidden-target"
	ReasonModerationMissingSignature PolicyReason = "moderation-missing-signature"
	ReasonModerationInvalidSigner    PolicyReason = "moderation-invalid-signer"
	ReasonModerationInvalidSignature PolicyReason = "moderation-invalid-signature"
	ReasonModerationNonCompliant     PolicyReason = "moderation-non-compliant"
)

type PolicyDecision struct {
	Allowed bool
	Reason  PolicyReason
}

func AuthorizeMention(role rbac.Role, mention MentionType) PolicyDecision {
	switch mention {
	case MentionTypeEveryone, MentionTypeHere, MentionTypeRole:
		if role == rbac.RoleOwner || role == rbac.RoleAdmin || role == rbac.RoleModerator {
			return PolicyDecision{Allowed: true, Reason: ReasonMentionAuthorized}
		}
		return PolicyDecision{Allowed: false, Reason: ReasonMentionUnauthorized}
	default:
		return PolicyDecision{Allowed: true, Reason: ReasonMentionAuthorized}
	}
}

type SlowModePolicy struct {
	Interval time.Duration
}

type SlowModeState map[string]time.Time

type SlowModeDecision struct {
	Allowed   bool
	Reason    PolicyReason
	NextState SlowModeState
}

type SlowModeReplayEvent struct {
	EventID string
	ActorID string
	SentAt  time.Time
}

type SlowModeReplayState struct {
	Activity SlowModeState
	Seen     map[string]time.Time
}

type SlowModeReplayDecision struct {
	Allowed   bool
	Reason    PolicyReason
	NextState SlowModeReplayState
}

func EvaluateSlowMode(policy SlowModePolicy, state SlowModeState, actorID string, now time.Time, bypass bool) SlowModeDecision {
	next := make(SlowModeState, len(state)+1)
	for k, v := range state {
		next[k] = v
	}
	if bypass {
		next[actorID] = now
		return SlowModeDecision{Allowed: true, Reason: ReasonSlowModeBypass, NextState: next}
	}
	if policy.Interval <= 0 {
		next[actorID] = now
		return SlowModeDecision{Allowed: true, Reason: ReasonSlowModePass, NextState: next}
	}
	last, ok := state[actorID]
	if ok && now.Sub(last) < policy.Interval {
		return SlowModeDecision{Allowed: false, Reason: ReasonSlowModeActive, NextState: next}
	}
	next[actorID] = now
	return SlowModeDecision{Allowed: true, Reason: ReasonSlowModePass, NextState: next}
}

func EvaluateSlowModeReplay(policy SlowModePolicy, state SlowModeReplayState, event SlowModeReplayEvent, bypass bool) SlowModeReplayDecision {
	next := copySlowModeReplayState(state)
	if event.EventID == "" || event.ActorID == "" || event.SentAt.IsZero() {
		return SlowModeReplayDecision{Allowed: false, Reason: ReasonSlowModeReplayStaleEvent, NextState: next}
	}
	if _, ok := next.Seen[event.EventID]; ok {
		return SlowModeReplayDecision{Allowed: false, Reason: ReasonSlowModeReplayDuplicate, NextState: next}
	}
	if !bypass {
		if last, ok := next.Activity[event.ActorID]; ok && !event.SentAt.After(last) {
			return SlowModeReplayDecision{Allowed: false, Reason: ReasonSlowModeReplayStaleEvent, NextState: next}
		}
	}
	decision := EvaluateSlowMode(policy, next.Activity, event.ActorID, event.SentAt, bypass)
	next.Activity = decision.NextState
	if !decision.Allowed {
		return SlowModeReplayDecision{Allowed: false, Reason: decision.Reason, NextState: next}
	}
	next.Seen[event.EventID] = event.SentAt
	return SlowModeReplayDecision{Allowed: true, Reason: decision.Reason, NextState: next}
}

func ReconcileSlowModeReplayState(local, remote SlowModeReplayState) SlowModeReplayState {
	merged := copySlowModeReplayState(local)
	for actor, ts := range remote.Activity {
		if existing, ok := merged.Activity[actor]; !ok || ts.After(existing) {
			merged.Activity[actor] = ts
		}
	}
	for eventID, ts := range remote.Seen {
		if existing, ok := merged.Seen[eventID]; !ok || ts.After(existing) {
			merged.Seen[eventID] = ts
		}
	}
	return merged
}

func copySlowModeReplayState(state SlowModeReplayState) SlowModeReplayState {
	activity := make(SlowModeState, len(state.Activity))
	for actor, ts := range state.Activity {
		activity[actor] = ts
	}
	seen := make(map[string]time.Time, len(state.Seen))
	for id, ts := range state.Seen {
		seen[id] = ts
	}
	return SlowModeReplayState{Activity: activity, Seen: seen}
}

type ModerationAction string

const (
	ActionRedact  ModerationAction = "redact"
	ActionDelete  ModerationAction = "delete"
	ActionTimeout ModerationAction = "timeout"
	ActionBan     ModerationAction = "ban"
)

type SignatureAlgorithm string

const (
	SignatureAlgorithmUnspecified SignatureAlgorithm = "unspecified"
	SignatureAlgorithmEd25519     SignatureAlgorithm = "ed25519"
)

type ModerationEventRecord struct {
	EventID   string
	ActorID   string
	TargetID  string
	Action    ModerationAction
	SignerID  string
	Algorithm SignatureAlgorithm
	Signature []byte
}

type SignatureVerifier func(event ModerationEventRecord) bool

func CanApplyModeration(actorRole, targetRole rbac.Role, action ModerationAction) PolicyDecision {
	_ = action
	if !rbac.CanActOnTarget(actorRole, targetRole) {
		return PolicyDecision{Allowed: false, Reason: ReasonModerationForbiddenTarget}
	}
	return PolicyDecision{Allowed: true, Reason: ReasonModerationAllowed}
}

func EnforceModerationEvent(actorRole, targetRole rbac.Role, event ModerationEventRecord, compliantClient bool, verify SignatureVerifier) PolicyDecision {
	decision := CanApplyModeration(actorRole, targetRole, event.Action)
	if !decision.Allowed {
		return decision
	}
	if !compliantClient {
		return PolicyDecision{Allowed: false, Reason: ReasonModerationNonCompliant}
	}
	if event.EventID == "" || event.ActorID == "" || event.TargetID == "" {
		return PolicyDecision{Allowed: false, Reason: ReasonModerationMissingSignature}
	}
	if event.SignerID == "" || event.Algorithm == SignatureAlgorithmUnspecified || len(event.Signature) == 0 {
		return PolicyDecision{Allowed: false, Reason: ReasonModerationMissingSignature}
	}
	if event.SignerID != event.ActorID {
		return PolicyDecision{Allowed: false, Reason: ReasonModerationInvalidSigner}
	}
	if verify != nil && !verify(event) {
		return PolicyDecision{Allowed: false, Reason: ReasonModerationInvalidSignature}
	}
	return PolicyDecision{Allowed: true, Reason: ReasonModerationAllowed}
}
