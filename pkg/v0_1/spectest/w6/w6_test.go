// Package w6 tests the W6 protocol families: chat, manifest, friends, notify, sync, presence,
// moderation, governance, and voice integration scenarios.
package w6

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	chatpkg "github.com/aether/code_aether/pkg/v0_1/family/chat"
	friendspkg "github.com/aether/code_aether/pkg/v0_1/family/friends"
	govpkg "github.com/aether/code_aether/pkg/v0_1/family/governance"
	manifestpkg "github.com/aether/code_aether/pkg/v0_1/family/manifest"
	modpkg "github.com/aether/code_aether/pkg/v0_1/family/moderation"
	notifypkg "github.com/aether/code_aether/pkg/v0_1/family/notify"
	presencepkg "github.com/aether/code_aether/pkg/v0_1/family/presence"
	syncpkg "github.com/aether/code_aether/pkg/v0_1/family/sync"
	voicepkg "github.com/aether/code_aether/pkg/v0_1/family/voice"
	mspkg "github.com/aether/code_aether/pkg/v0_1/mode/mediashield"
	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
)

func mustMarshal(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

// chatReq builds a PeerStreamRequest advertising cap.chat.
func chatReq(op string, payload any) *proto.PeerStreamRequest {
	return &proto.PeerStreamRequest{
		Operation:     op,
		AdvertisedCaps: []string{"cap.chat"},
		Payload:       mustMarshal(payload),
	}
}

// manifestReq builds a PeerStreamRequest advertising cap.manifest.
func manifestReq(op string, payload any) *proto.PeerStreamRequest {
	return &proto.PeerStreamRequest{
		Operation:     op,
		AdvertisedCaps: []string{"cap.manifest"},
		Payload:       mustMarshal(payload),
	}
}

// --- Chat ---

func TestW6ChatSendAndHistory(t *testing.T) {
	h := &chatpkg.Handler{}
	ctx := context.Background()

	// Join a scope.
	joinResp := h.HandleStream(ctx, chatReq("chat.join", map[string]any{"scope_id": "ch1", "peer_id": "alice"}))
	if joinResp.Error != nil {
		t.Fatalf("chat.join: %s", joinResp.Error.Message)
	}

	// Send a message.
	sendResp := h.HandleStream(ctx, chatReq("chat.send", map[string]any{
		"scope_id":  "ch1",
		"sender_id": "alice",
		"body":      []byte("hello world"),
	}))
	if sendResp.Error != nil {
		t.Fatalf("chat.send: %s", sendResp.Error.Message)
	}

	// Fetch history.
	histResp := h.HandleStream(ctx, chatReq("chat.history", map[string]any{"scope_id": "ch1"}))
	if histResp.Error != nil {
		t.Fatalf("chat.history: %s", histResp.Error.Message)
	}

	var result struct {
		Messages []struct {
			Body []byte `json:"Body"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(histResp.Payload, &result); err != nil {
		t.Fatalf("unmarshal history: %v", err)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}
}

func TestW6ChatHistoryLimit(t *testing.T) {
	h := &chatpkg.Handler{}
	ctx := context.Background()

	h.HandleStream(ctx, chatReq("chat.join", map[string]any{"scope_id": "ch2", "peer_id": "alice"}))
	for i := 0; i < 5; i++ {
		h.HandleStream(ctx, chatReq("chat.send", map[string]any{
			"scope_id":  "ch2",
			"sender_id": "alice",
			"body":      []byte("msg"),
		}))
	}

	histResp := h.HandleStream(ctx, chatReq("chat.history", map[string]any{"scope_id": "ch2", "limit": 3}))
	var result struct {
		Messages []struct{} `json:"messages"`
	}
	json.Unmarshal(histResp.Payload, &result)
	if len(result.Messages) != 3 {
		t.Fatalf("expected 3 messages with limit=3, got %d", len(result.Messages))
	}
}

func TestW6ChatUnknownOp(t *testing.T) {
	h := &chatpkg.Handler{}
	resp := h.HandleStream(context.Background(), &proto.PeerStreamRequest{
		Operation:     "chat.noop",
		AdvertisedCaps: []string{"cap.chat"},
		Payload:       mustMarshal(map[string]any{}),
	})
	if resp.Error == nil || resp.Error.Code != proto.CodeUnsupportedOperation {
		t.Fatalf("expected UNSUPPORTED_OPERATION, got %v", resp.Error)
	}
}

// --- Manifest ---

func TestW6ManifestPublishAndFetch(t *testing.T) {
	h := &manifestpkg.Handler{}
	ctx := context.Background()

	content := json.RawMessage(`{"key":"value"}`)
	pubResp := h.HandleStream(ctx, manifestReq("manifest.publish", map[string]any{
		"scope_id":     "sc1",
		"publisher_id": "alice",
		"content":      content,
	}))
	if pubResp.Error != nil {
		t.Fatalf("manifest.publish: %s", pubResp.Error.Message)
	}

	var pubResult struct {
		Hash string `json:"hash"`
	}
	json.Unmarshal(pubResp.Payload, &pubResult)
	if pubResult.Hash == "" {
		t.Fatal("expected non-empty hash")
	}

	// Fetch by hash.
	fetchResp := h.HandleStream(ctx, manifestReq("manifest.fetch", map[string]any{"hash": pubResult.Hash}))
	if fetchResp.Error != nil {
		t.Fatalf("manifest.fetch by hash: %s", fetchResp.Error.Message)
	}

	var m manifestpkg.ManifestRecord
	if err := json.Unmarshal(fetchResp.Payload, &m); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	if m.Hash != pubResult.Hash {
		t.Fatalf("hash mismatch: %s vs %s", m.Hash, pubResult.Hash)
	}
}

func TestW6ManifestFetchByScope(t *testing.T) {
	h := &manifestpkg.Handler{}
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		content := mustMarshal(map[string]any{"i": i})
		h.HandleStream(ctx, manifestReq("manifest.publish", map[string]any{
			"scope_id":     "scope-a",
			"publisher_id": "alice",
			"content":      json.RawMessage(content),
		}))
	}
	// Publish one with a different scope.
	h.HandleStream(ctx, manifestReq("manifest.publish", map[string]any{
		"scope_id":     "scope-b",
		"publisher_id": "alice",
		"content":      json.RawMessage(`{"other":true}`),
	}))

	fetchResp := h.HandleStream(ctx, manifestReq("manifest.fetch", map[string]any{"scope_id": "scope-a"}))
	var result struct {
		Manifests []manifestpkg.ManifestRecord `json:"manifests"`
	}
	json.Unmarshal(fetchResp.Payload, &result)
	if len(result.Manifests) != 3 {
		t.Fatalf("expected 3 manifests for scope-a, got %d", len(result.Manifests))
	}
}

// --- Presence ---

func TestW6PresenceUpdateAndQuery(t *testing.T) {
	h := &presencepkg.Handler{}
	ctx := context.Background()

	h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "presence.update",
		Payload:   mustMarshal(map[string]any{"peer_id": "alice", "status": "online"}),
	})
	h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "presence.update",
		Payload:   mustMarshal(map[string]any{"peer_id": "bob", "status": "away"}),
	})

	queryResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "presence.query",
		Payload:   mustMarshal(map[string]any{"peer_ids": []string{"alice", "bob", "carol"}}),
	})
	if queryResp.Error != nil {
		t.Fatalf("presence.query: %s", queryResp.Error.Message)
	}

	var result struct {
		Presence []presencepkg.PresenceRecord `json:"presence"`
	}
	json.Unmarshal(queryResp.Payload, &result)
	if len(result.Presence) != 2 {
		t.Fatalf("expected 2 records (carol unknown), got %d", len(result.Presence))
	}
}

// --- Notify ---

func TestW6NotifyPushAndDrain(t *testing.T) {
	h := &notifypkg.Handler{}
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		h.HandleStream(ctx, &proto.PeerStreamRequest{
			Operation: "notify.push",
			Payload:   mustMarshal(map[string]any{"recipient_id": "alice", "type": "msg", "payload": json.RawMessage(`{}`)}),
		})
	}

	drainResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "notify.drain",
		Payload:   mustMarshal(map[string]any{"recipient_id": "alice"}),
	})
	if drainResp.Error != nil {
		t.Fatalf("notify.drain: %s", drainResp.Error.Message)
	}

	var result struct {
		Count int `json:"count"`
	}
	json.Unmarshal(drainResp.Payload, &result)
	if result.Count != 3 {
		t.Fatalf("expected 3 notifications, got %d", result.Count)
	}

	// Second drain: empty.
	drainResp2 := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "notify.drain",
		Payload:   mustMarshal(map[string]any{"recipient_id": "alice"}),
	})
	var result2 struct {
		Count int `json:"count"`
	}
	json.Unmarshal(drainResp2.Payload, &result2)
	if result2.Count != 0 {
		t.Fatalf("expected 0 notifications after drain, got %d", result2.Count)
	}
}

func TestW6NotifyDirectPushDrain(t *testing.T) {
	h := &notifypkg.Handler{}
	h.Push("bob", "friend_request", json.RawMessage(`{"from":"alice"}`))
	h.Push("bob", "friend_request", json.RawMessage(`{"from":"carol"}`))
	ns := h.Drain("bob")
	if len(ns) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(ns))
	}
	if len(h.Drain("bob")) != 0 {
		t.Fatal("second drain should be empty")
	}
}

// --- Friends ---

func TestW6FriendsRequestAcceptList(t *testing.T) {
	h := &friendspkg.Handler{}
	ctx := context.Background()

	reqResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "friends.request",
		Payload:   mustMarshal(map[string]any{"from_peer_id": "alice", "to_peer_id": "bob"}),
	})
	if reqResp.Error != nil {
		t.Fatalf("friends.request: %s", reqResp.Error.Message)
	}

	accResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "friends.accept",
		Payload:   mustMarshal(map[string]any{"from_peer_id": "alice", "to_peer_id": "bob"}),
	})
	if accResp.Error != nil {
		t.Fatalf("friends.accept: %s", accResp.Error.Message)
	}

	listResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "friends.list",
		Payload:   mustMarshal(map[string]any{"peer_id": "alice"}),
	})
	var result struct {
		Friends []string `json:"friends"`
	}
	json.Unmarshal(listResp.Payload, &result)
	if len(result.Friends) != 1 || result.Friends[0] != "bob" {
		t.Fatalf("expected alice's friends = [bob], got %v", result.Friends)
	}

	// Bidirectional: bob's friends should also contain alice.
	listResp2 := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "friends.list",
		Payload:   mustMarshal(map[string]any{"peer_id": "bob"}),
	})
	var result2 struct {
		Friends []string `json:"friends"`
	}
	json.Unmarshal(listResp2.Payload, &result2)
	if len(result2.Friends) != 1 || result2.Friends[0] != "alice" {
		t.Fatalf("expected bob's friends = [alice], got %v", result2.Friends)
	}
}

func TestW6FriendsRateLimit(t *testing.T) {
	h := &friendspkg.Handler{}
	ctx := context.Background()

	// Send 10 requests (max allowed per hour).
	for i := 0; i < 10; i++ {
		resp := h.HandleStream(ctx, &proto.PeerStreamRequest{
			Operation: "friends.request",
			Payload:   mustMarshal(map[string]any{"from_peer_id": "spammer", "to_peer_id": "victim"}),
		})
		if resp.Error != nil {
			t.Fatalf("request %d unexpected error: %s", i, resp.Error.Message)
		}
	}

	// 11th must be rate-limited.
	resp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "friends.request",
		Payload:   mustMarshal(map[string]any{"from_peer_id": "spammer", "to_peer_id": "victim2"}),
	})
	if resp.Error == nil || resp.Error.Code != proto.CodeRateLimited {
		t.Fatalf("expected RATE_LIMITED on 11th request, got %v", resp.Error)
	}
}

func TestW6FriendsAcceptExpired(t *testing.T) {
	h := &friendspkg.Handler{}
	ctx := context.Background()

	h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "friends.request",
		Payload:   mustMarshal(map[string]any{"from_peer_id": "alice", "to_peer_id": "bob"}),
	})

	// Manually expire the request.
	h.ExpireRequestForTest("alice", "bob", time.Now().Add(-8*24*time.Hour))

	accResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "friends.accept",
		Payload:   mustMarshal(map[string]any{"from_peer_id": "alice", "to_peer_id": "bob"}),
	})
	if accResp.Error == nil {
		t.Fatal("expected error for expired request")
	}
}

func TestW6FriendsAcceptNonexistent(t *testing.T) {
	h := &friendspkg.Handler{}
	ctx := context.Background()

	resp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation: "friends.accept",
		Payload:   mustMarshal(map[string]any{"from_peer_id": "nobody", "to_peer_id": "ghost"}),
	})
	if resp.Error == nil {
		t.Fatal("expected error for non-existent request")
	}
}

// --- Sync ---

func TestW6SyncCoverageFetch(t *testing.T) {
	h := syncpkg.NewHandler(false)
	ctx := context.Background()
	h.AddMember("srv1", "peer1")

	ts := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	h.SeedMessageForTest(&syncpkg.Message{
		ID: "msg-1", ServerID: "srv1", Sequence: 1, CreatedAt: ts,
		Body: []byte("hello"), Signature: []byte("sig1"),
	})
	h.SeedMessageForTest(&syncpkg.Message{
		ID: "msg-2", ServerID: "srv1", Sequence: 2, CreatedAt: ts.Add(time.Minute),
		Body: []byte("world"), Signature: []byte("sig2"),
	})

	// Coverage: both messages.
	covResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "sync.coverage",
		AdvertisedCaps: []string{"cap.sync"},
		Payload:        mustMarshal(map[string]any{"server_id": "srv1", "from_seq": 1, "to_seq": 2}),
	})
	if covResp.Error != nil {
		t.Fatalf("coverage: %s", covResp.Error.Message)
	}
	var covResult struct {
		AvailableFrom int64  `json:"available_from"`
		AvailableTo   int64  `json:"available_to"`
		SnapshotRoot  string `json:"snapshot_root"`
	}
	json.Unmarshal(covResp.Payload, &covResult)
	if covResult.AvailableFrom != 1 || covResult.AvailableTo != 2 {
		t.Errorf("coverage range: want 1..2, got %d..%d", covResult.AvailableFrom, covResult.AvailableTo)
	}
	if covResult.SnapshotRoot == "" {
		t.Error("expected non-empty snapshot_root")
	}

	// Fetch both messages.
	fetchResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "sync.fetch",
		AdvertisedCaps: []string{"cap.sync"},
		Payload: mustMarshal(map[string]any{
			"server_id":        "srv1",
			"requester_peer_id": "peer1",
			"message_ids":      []string{"msg-1", "msg-2"},
		}),
	})
	if fetchResp.Error != nil {
		t.Fatalf("fetch: %s", fetchResp.Error.Message)
	}
	var fetchResult struct {
		Messages    []map[string]any `json:"messages"`
		NotFoundIDs []string         `json:"not_found_ids"`
	}
	json.Unmarshal(fetchResp.Payload, &fetchResult)
	if len(fetchResult.Messages) != 2 {
		t.Errorf("fetch: want 2 messages, got %d", len(fetchResult.Messages))
	}
}

func TestW6SyncCoverageEmpty(t *testing.T) {
	h := syncpkg.NewHandler(false)
	resp := h.HandleStream(context.Background(), &proto.PeerStreamRequest{
		Operation:      "sync.coverage",
		AdvertisedCaps: []string{"cap.sync"},
		Payload:        mustMarshal(map[string]any{"server_id": "unknown", "from_seq": 0, "to_seq": 0}),
	})
	if resp.Error != nil {
		t.Fatalf("unexpected error: %s", resp.Error.Message)
	}
	var result struct {
		AvailableFrom int64  `json:"available_from"`
		AvailableTo   int64  `json:"available_to"`
		SnapshotRoot  string `json:"snapshot_root"`
	}
	json.Unmarshal(resp.Payload, &result)
	if result.AvailableFrom != 0 || result.AvailableTo != 0 || result.SnapshotRoot != "" {
		t.Errorf("expected empty coverage, got from=%d to=%d root=%q",
			result.AvailableFrom, result.AvailableTo, result.SnapshotRoot)
	}
}

// --- Governance integration ---

func TestW6GovernanceAssignRevokeSync(t *testing.T) {
	gov := govpkg.NewHandler()
	gov.AssignForTest("srv1", "owner1", govpkg.RoleOwner)
	gov.AssignForTest("srv1", "admin1", govpkg.RoleAdmin)
	gov.AssignForTest("srv1", "mod1", govpkg.RoleModerator)
	gov.AssignForTest("srv1", "member1", govpkg.RoleMember)
	gov.AssignForTest("srv1", "member2", govpkg.RoleMember)
	ctx := context.Background()

	// Owner assigns member1 as admin (requires owner).
	assignResp := gov.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "governance.assign_role",
		AdvertisedCaps: []string{"cap.rbac"},
		Payload: mustMarshal(map[string]any{
			"actor_peer_id":  "owner1",
			"target_peer_id": "member1",
			"server_id":      "srv1",
			"role":           "admin",
			"policy_version": 0,
		}),
	})
	if assignResp.Error != nil {
		t.Fatalf("assign admin: %s", assignResp.Error.Message)
	}

	// Moderator attempts assign_admin on member2 — must fail (unauthorized; moderators can't assign admin).
	failResp := gov.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "governance.assign_role",
		AdvertisedCaps: []string{"cap.rbac"},
		Payload: mustMarshal(map[string]any{
			"actor_peer_id":  "mod1",
			"target_peer_id": "member2",
			"server_id":      "srv1",
			"role":           "admin",
			"policy_version": 1,
		}),
	})
	if failResp.Error == nil || failResp.Error.Code != proto.CodeGovernanceUnauthorized {
		t.Fatalf("expected GOVERNANCE_UNAUTHORIZED, got %v", failResp.Error)
	}

	// Owner revokes admin1.
	revokeResp := gov.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "governance.revoke_role",
		AdvertisedCaps: []string{"cap.rbac"},
		Payload: mustMarshal(map[string]any{
			"actor_peer_id":  "owner1",
			"target_peer_id": "admin1",
			"server_id":      "srv1",
			"role":           "member",
			"policy_version": 1,
		}),
	})
	if revokeResp.Error != nil {
		t.Fatalf("revoke admin: %s", revokeResp.Error.Message)
	}

	// Sync — new peer with policy_version=0 should get full table.
	syncResp := gov.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "governance.sync",
		AdvertisedCaps: []string{"cap.rbac"},
		Payload: mustMarshal(map[string]any{
			"requester_peer_id":    "newcomer",
			"server_id":           "srv1",
			"known_policy_version": 0,
		}),
	})
	if syncResp.Error != nil {
		t.Fatalf("sync: %s", syncResp.Error.Message)
	}
	var syncResult struct {
		UpToDate bool `json:"up_to_date"`
	}
	json.Unmarshal(syncResp.Payload, &syncResult)
	if syncResult.UpToDate {
		t.Error("expected up_to_date=false for newcomer with policy_version=0")
	}
}

// --- Moderation integration ---

func TestW6ModerationKickBanMuteDeleteMessage(t *testing.T) {
	gov := govpkg.NewHandler()
	gov.AssignForTest("srv1", "admin1", govpkg.RoleAdmin)
	gov.AssignForTest("srv1", "mod1", govpkg.RoleModerator)
	gov.AssignForTest("srv1", "member1", govpkg.RoleMember)
	gov.AssignForTest("srv1", "member2", govpkg.RoleMember)
	mod := modpkg.New(gov)
	mod.AddMember("srv1", "member1")
	mod.AddMember("srv1", "member2")
	mod.StoreMessage("msg-001")
	ctx := context.Background()
	caps := []string{"cap.moderation", "cap.slow-mode"}

	// Moderator kicks member1 — success.
	kickResp := mod.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "moderation.kick",
		AdvertisedCaps: caps,
		Payload: mustMarshal(map[string]any{
			"actor_peer_id":  "mod1",
			"target_peer_id": "member1",
			"scope_id":       "srv1",
		}),
	})
	if kickResp.Error != nil {
		t.Fatalf("kick: %s", kickResp.Error.Message)
	}

	// Admin bans member2 — success.
	banResp := mod.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "moderation.ban",
		AdvertisedCaps: caps,
		Payload: mustMarshal(map[string]any{
			"actor_peer_id":  "admin1",
			"target_peer_id": "member2",
			"scope_id":       "srv1",
		}),
	})
	if banResp.Error != nil {
		t.Fatalf("ban: %s", banResp.Error.Message)
	}

	// Moderator mutes member2 for 5000ms — success.
	muteResp := mod.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "moderation.mute",
		AdvertisedCaps: caps,
		Payload: mustMarshal(map[string]any{
			"actor_peer_id":  "mod1",
			"target_peer_id": "member2",
			"scope_id":       "srv1",
			"duration_ms":    5000,
		}),
	})
	if muteResp.Error != nil {
		t.Fatalf("mute: %s", muteResp.Error.Message)
	}
	if !mod.IsMuted("srv1", "member2") {
		t.Error("expected member2 to be muted")
	}

	// Moderator deletes msg-001 — success.
	delResp := mod.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "moderation.delete_message",
		AdvertisedCaps: caps,
		Payload: mustMarshal(map[string]any{
			"actor_peer_id": "mod1",
			"message_id":    "msg-001",
			"server_id":     "srv1",
		}),
	})
	if delResp.Error != nil {
		t.Fatalf("delete_message: %s", delResp.Error.Message)
	}

	// Member1 attempts kick (insufficient role) — must fail.
	unauthorized := mod.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "moderation.kick",
		AdvertisedCaps: caps,
		Payload: mustMarshal(map[string]any{
			"actor_peer_id":  "member1",
			"target_peer_id": "member2",
			"scope_id":       "srv1",
		}),
	})
	if unauthorized.Error == nil || unauthorized.Error.Code != proto.CodeModerationUnauthorized {
		t.Fatalf("expected MODERATION_UNAUTHORIZED, got %v", unauthorized.Error)
	}
}

// --- Voice integration ---

func TestW6VoiceJoinFrameLeave(t *testing.T) {
	h := &voicepkg.Handler{LocalPeerID: "server"}
	keys := map[string]*mspkg.PeerKey{
		"alice": {PeerID: "alice", Key: make([]byte, 32)},
		"bob":   {PeerID: "bob", Key: make([]byte, 32)},
	}
	if err := h.CreateSession("session1", []string{"alice", "bob"}, keys); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	ctx := context.Background()
	caps := []string{"cap.voice", "mode.mediashield"}

	// Join.
	joinResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "voice.join",
		AdvertisedCaps: caps,
		Payload:        mustMarshal(map[string]any{"session_id": "session1", "peer_id": "alice"}),
	})
	if joinResp.Error != nil {
		t.Fatalf("join: %s", joinResp.Error.Message)
	}

	// SDP offer with Opus.
	offerResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "voice.offer",
		AdvertisedCaps: caps,
		Payload: mustMarshal(map[string]any{
			"session_id": "session1",
			"peer_id":    "alice",
			"sdp":        "v=0\r\na=rtpmap:111 opus/48000/2\r\n",
			"sequence":   1,
			"expires_at": 0,
		}),
	})
	if offerResp.Error != nil {
		t.Fatalf("offer: %s", offerResp.Error.Message)
	}

	// Frame forwarded opaquely — counter=1.
	frameResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "voice.frame",
		AdvertisedCaps: caps,
		Payload: mustMarshal(map[string]any{
			"session_id": "session1",
			"sender_id":  "alice",
			"ciphertext": []byte{0xAB, 0xCD},
			"counter":    1,
		}),
	})
	if frameResp.Error != nil {
		t.Fatalf("frame: %s", frameResp.Error.Message)
	}

	// Replay same counter — must fail.
	replayResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "voice.frame",
		AdvertisedCaps: caps,
		Payload: mustMarshal(map[string]any{
			"session_id": "session1",
			"sender_id":  "alice",
			"ciphertext": []byte{0xAB, 0xCD},
			"counter":    1,
		}),
	})
	if replayResp.Error == nil || replayResp.Error.Code != proto.CodeReplayDetected {
		t.Fatalf("expected REPLAY_DETECTED, got %v", replayResp.Error)
	}

	// Terminate session.
	termResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "voice.terminate",
		AdvertisedCaps: caps,
		Payload:        mustMarshal(map[string]any{"session_id": "session1", "peer_id": "alice"}),
	})
	if termResp.Error != nil {
		t.Fatalf("terminate: %s", termResp.Error.Message)
	}

	// Frame after terminate — must fail.
	afterResp := h.HandleStream(ctx, &proto.PeerStreamRequest{
		Operation:      "voice.frame",
		AdvertisedCaps: caps,
		Payload: mustMarshal(map[string]any{
			"session_id": "session1",
			"sender_id":  "alice",
			"ciphertext": []byte{0xAB, 0xCD},
			"counter":    2,
		}),
	})
	if afterResp.Error == nil || afterResp.Error.Code != proto.CodeVoiceSessionNotFound {
		t.Fatalf("expected VOICE_SESSION_NOT_FOUND, got %v", afterResp.Error)
	}
}
