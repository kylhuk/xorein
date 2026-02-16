package policy

import (
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v02/rbac"
)

func TestParseMentions(t *testing.T) {
	tokens := ParseMentions(`hello @user and @here but not \\@escaped`)
	if len(tokens) != 2 {
		t.Fatalf("len(tokens)=%d want 2", len(tokens))
	}
	if tokens[0].Type != MentionTypeUser || tokens[1].Type != MentionTypeHere {
		t.Fatalf("unexpected tokens: %#v", tokens)
	}
}

func TestResolveMentionToken(t *testing.T) {
	directory := MentionDirectory{
		Users: map[string]MentionTargetStatus{
			"alice": MentionTargetValid,
			"dual":  MentionTargetStale,
			"stale": MentionTargetStale,
		},
		Roles: map[string]MentionTargetStatus{
			"team":  MentionTargetValid,
			"dual":  MentionTargetValid,
			"ghost": MentionTargetStale,
		},
	}

	tests := []struct {
		name       string
		token      MentionToken
		wantType   MentionType
		wantTarget string
		wantStatus MentionTargetStatus
		wantNotify bool
		wantRender string
	}{
		{
			name:       "valid user mention resolves to user",
			token:      MentionToken{Type: MentionTypeUser, Raw: "@alice"},
			wantType:   MentionTypeUser,
			wantTarget: "alice",
			wantStatus: MentionTargetValid,
			wantNotify: true,
			wantRender: "@alice",
		},
		{
			name:       "user mention falls back to valid role",
			token:      MentionToken{Type: MentionTypeUser, Raw: "@team"},
			wantType:   MentionTypeRole,
			wantTarget: "team",
			wantStatus: MentionTargetValid,
			wantNotify: true,
			wantRender: "@team",
		},
		{
			name:       "valid role wins over stale user",
			token:      MentionToken{Type: MentionTypeUser, Raw: "@dual"},
			wantType:   MentionTypeRole,
			wantTarget: "dual",
			wantStatus: MentionTargetValid,
			wantNotify: true,
			wantRender: "@dual",
		},
		{
			name:       "stale user suppresses notification",
			token:      MentionToken{Type: MentionTypeUser, Raw: "@stale"},
			wantType:   MentionTypeUser,
			wantTarget: "stale",
			wantStatus: MentionTargetStale,
			wantNotify: false,
			wantRender: "@stale",
		},
		{
			name:       "stale role suppresses notification",
			token:      MentionToken{Type: MentionTypeRole, Raw: "@ghost"},
			wantType:   MentionTypeRole,
			wantTarget: "ghost",
			wantStatus: MentionTargetStale,
			wantNotify: false,
			wantRender: "@ghost",
		},
		{
			name:       "unknown target falls back to raw token",
			token:      MentionToken{Type: MentionTypeUser, Raw: "@missing,"},
			wantType:   MentionTypeUser,
			wantTarget: "missing",
			wantStatus: MentionTargetUnknown,
			wantNotify: false,
			wantRender: "@missing,",
		},
		{
			name:       "here mention always resolves",
			token:      MentionToken{Type: MentionTypeHere, Raw: "@here"},
			wantType:   MentionTypeHere,
			wantTarget: "here",
			wantStatus: MentionTargetValid,
			wantNotify: true,
			wantRender: "@here",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveMentionToken(tc.token, directory)
			if got.Type != tc.wantType || got.Target != tc.wantTarget || got.Status != tc.wantStatus || got.Notify != tc.wantNotify || got.Render != tc.wantRender {
				t.Fatalf("ResolveMentionToken(%+v)=%+v want type=%q target=%q status=%q notify=%v render=%q", tc.token, got, tc.wantType, tc.wantTarget, tc.wantStatus, tc.wantNotify, tc.wantRender)
			}
		})
	}
}

func TestAuthorizeMention(t *testing.T) {
	allowed := AuthorizeMention(rbac.RoleModerator, MentionTypeEveryone)
	if !allowed.Allowed || allowed.Reason != ReasonMentionAuthorized {
		t.Fatalf("moderator authorize mismatch: %+v", allowed)
	}
	denied := AuthorizeMention(rbac.RoleMember, MentionTypeEveryone)
	if denied.Allowed || denied.Reason != ReasonMentionUnauthorized {
		t.Fatalf("member authorize mismatch: %+v", denied)
	}
}

func TestEvaluateSlowMode(t *testing.T) {
	now := time.Unix(100, 0)
	policy := SlowModePolicy{Interval: 5 * time.Second}
	state := SlowModeState{}
	first := EvaluateSlowMode(policy, state, "member", now, false)
	if !first.Allowed {
		t.Fatalf("first send should pass: %+v", first)
	}
	state = first.NextState
	second := EvaluateSlowMode(policy, state, "member", now.Add(time.Second), false)
	if second.Allowed || second.Reason != ReasonSlowModeActive {
		t.Fatalf("second send should be throttled: %+v", second)
	}
	bypass := EvaluateSlowMode(policy, state, "owner", now.Add(time.Second), true)
	if !bypass.Allowed || bypass.Reason != ReasonSlowModeBypass {
		t.Fatalf("owner bypass mismatch: %+v", bypass)
	}
}

func TestCanApplyModeration(t *testing.T) {
	allowed := CanApplyModeration(rbac.RoleAdmin, rbac.RoleMember, ActionTimeout)
	if !allowed.Allowed {
		t.Fatalf("admin->member timeout should pass: %+v", allowed)
	}
	denied := CanApplyModeration(rbac.RoleModerator, rbac.RoleAdmin, ActionBan)
	if denied.Allowed || denied.Reason != ReasonModerationForbiddenTarget {
		t.Fatalf("moderator->admin ban should fail: %+v", denied)
	}
}

func TestEnforceModerationEvent(t *testing.T) {
	base := ModerationEventRecord{
		EventID:   "evt-1",
		ActorID:   "admin",
		TargetID:  "member",
		Action:    ActionTimeout,
		SignerID:  "admin",
		Algorithm: SignatureAlgorithmEd25519,
		Signature: []byte{1, 2, 3},
	}

	tests := []struct {
		name       string
		actorRole  rbac.Role
		targetRole rbac.Role
		event      ModerationEventRecord
		compliant  bool
		verify     SignatureVerifier
		wantAllow  bool
		wantReason PolicyReason
	}{
		{
			name:       "valid signed moderation passes",
			actorRole:  rbac.RoleAdmin,
			targetRole: rbac.RoleMember,
			event:      base,
			compliant:  true,
			verify: func(event ModerationEventRecord) bool {
				return event.EventID == "evt-1"
			},
			wantAllow:  true,
			wantReason: ReasonModerationAllowed,
		},
		{
			name:       "missing signature rejected",
			actorRole:  rbac.RoleAdmin,
			targetRole: rbac.RoleMember,
			event: func() ModerationEventRecord {
				e := base
				e.Signature = nil
				return e
			}(),
			compliant:  true,
			verify:     nil,
			wantAllow:  false,
			wantReason: ReasonModerationMissingSignature,
		},
		{
			name:       "signer mismatch rejected",
			actorRole:  rbac.RoleAdmin,
			targetRole: rbac.RoleMember,
			event: func() ModerationEventRecord {
				e := base
				e.SignerID = "other"
				return e
			}(),
			compliant:  true,
			verify:     nil,
			wantAllow:  false,
			wantReason: ReasonModerationInvalidSigner,
		},
		{
			name:       "invalid signature callback rejected",
			actorRole:  rbac.RoleAdmin,
			targetRole: rbac.RoleMember,
			event:      base,
			compliant:  true,
			verify: func(ModerationEventRecord) bool {
				return false
			},
			wantAllow:  false,
			wantReason: ReasonModerationInvalidSignature,
		},
		{
			name:       "non compliant client rejected",
			actorRole:  rbac.RoleAdmin,
			targetRole: rbac.RoleMember,
			event:      base,
			compliant:  false,
			verify:     nil,
			wantAllow:  false,
			wantReason: ReasonModerationNonCompliant,
		},
		{
			name:       "forbidden actor target still rejected",
			actorRole:  rbac.RoleModerator,
			targetRole: rbac.RoleAdmin,
			event:      base,
			compliant:  true,
			verify:     nil,
			wantAllow:  false,
			wantReason: ReasonModerationForbiddenTarget,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := EnforceModerationEvent(tc.actorRole, tc.targetRole, tc.event, tc.compliant, tc.verify)
			if got.Allowed != tc.wantAllow || got.Reason != tc.wantReason {
				t.Fatalf("EnforceModerationEvent()=%+v want allowed=%v reason=%q", got, tc.wantAllow, tc.wantReason)
			}
		})
	}
}

func TestEvaluateSlowModeReplay(t *testing.T) {
	now := time.Unix(100, 0)
	policy := SlowModePolicy{Interval: 5 * time.Second}
	state := SlowModeReplayState{Activity: SlowModeState{}, Seen: map[string]time.Time{}}

	first := EvaluateSlowModeReplay(policy, state, SlowModeReplayEvent{EventID: "evt-1", ActorID: "member", SentAt: now}, false)
	if !first.Allowed || first.Reason != ReasonSlowModePass {
		t.Fatalf("first replay event should pass: %+v", first)
	}

	duplicate := EvaluateSlowModeReplay(policy, first.NextState, SlowModeReplayEvent{EventID: "evt-1", ActorID: "member", SentAt: now.Add(10 * time.Second)}, false)
	if duplicate.Allowed || duplicate.Reason != ReasonSlowModeReplayDuplicate {
		t.Fatalf("duplicate replay event should fail: %+v", duplicate)
	}

	stale := EvaluateSlowModeReplay(policy, first.NextState, SlowModeReplayEvent{EventID: "evt-2", ActorID: "member", SentAt: now.Add(-time.Second)}, false)
	if stale.Allowed || stale.Reason != ReasonSlowModeReplayStaleEvent {
		t.Fatalf("stale replay event should fail: %+v", stale)
	}

	throttled := EvaluateSlowModeReplay(policy, first.NextState, SlowModeReplayEvent{EventID: "evt-3", ActorID: "member", SentAt: now.Add(time.Second)}, false)
	if throttled.Allowed || throttled.Reason != ReasonSlowModeActive {
		t.Fatalf("throttled replay event should fail: %+v", throttled)
	}

	bypass := EvaluateSlowModeReplay(policy, first.NextState, SlowModeReplayEvent{EventID: "evt-4", ActorID: "owner", SentAt: now}, true)
	if !bypass.Allowed || bypass.Reason != ReasonSlowModeBypass {
		t.Fatalf("bypass replay event should pass: %+v", bypass)
	}
}

func TestReconcileSlowModeReplayState(t *testing.T) {
	baseTime := time.Unix(100, 0)
	local := SlowModeReplayState{
		Activity: SlowModeState{
			"member": baseTime,
			"admin":  baseTime.Add(5 * time.Second),
		},
		Seen: map[string]time.Time{
			"evt-1": baseTime,
		},
	}
	remote := SlowModeReplayState{
		Activity: SlowModeState{
			"member": baseTime.Add(10 * time.Second),
			"owner":  baseTime.Add(7 * time.Second),
		},
		Seen: map[string]time.Time{
			"evt-1": baseTime.Add(3 * time.Second),
			"evt-2": baseTime.Add(11 * time.Second),
		},
	}

	merged := ReconcileSlowModeReplayState(local, remote)
	if !merged.Activity["member"].Equal(baseTime.Add(10 * time.Second)) {
		t.Fatalf("member timestamp mismatch: %v", merged.Activity["member"])
	}
	if !merged.Activity["admin"].Equal(baseTime.Add(5 * time.Second)) {
		t.Fatalf("admin timestamp mismatch: %v", merged.Activity["admin"])
	}
	if !merged.Activity["owner"].Equal(baseTime.Add(7 * time.Second)) {
		t.Fatalf("owner timestamp mismatch: %v", merged.Activity["owner"])
	}
	if !merged.Seen["evt-1"].Equal(baseTime.Add(3 * time.Second)) {
		t.Fatalf("evt-1 timestamp mismatch: %v", merged.Seen["evt-1"])
	}
	if !merged.Seen["evt-2"].Equal(baseTime.Add(11 * time.Second)) {
		t.Fatalf("evt-2 timestamp mismatch: %v", merged.Seen["evt-2"])
	}
}
