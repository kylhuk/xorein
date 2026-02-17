package v07e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v07/archivist"
	"github.com/aether/code_aether/pkg/v07/history"
	"github.com/aether/code_aether/pkg/v07/historysync"
	"github.com/aether/code_aether/pkg/v07/merkle"
	"github.com/aether/code_aether/pkg/v07/notification"
	"github.com/aether/code_aether/pkg/v07/push"
	"github.com/aether/code_aether/pkg/v07/retention"
	"github.com/aether/code_aether/pkg/v07/search"
	"github.com/aether/code_aether/pkg/v07/storeforward"
)

func TestE2EScenarios(t *testing.T) {
	t.Run("E2E-SF-01", func(t *testing.T) {
		now := time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)
		policy := storeforward.TTLPolicy{Min: time.Minute, Max: 20 * time.Minute}
		record := storeforward.Record{
			ID:        "offline",
			Payload:   []byte("ciphertext:hello"),
			TTL:       10 * time.Minute,
			CreatedAt: now,
			Metadata:  map[string]string{},
		}
		if !policy.IsValid(record.TTL) {
			t.Fatalf("expected ttl to be valid")
		}
		resolved := retention.ResolveRetentionPolicy([]retention.ServerTier{retention.TierBurst, retention.TierEdge})
		if resolved.Tier != retention.TierEdge {
			t.Fatalf("unexpected resolved tier %s", resolved.Tier)
		}
		if retention.DetermineTransition(resolved, retention.RetentionPolicy{Tier: retention.TierBurst, Days: 7}) != retention.TransitionPurge {
			t.Fatalf("expected retention tightening to purge")
		}

		sendClassification := storeforward.ClassifyPurge(record, now, 0, 0)
		if sendClassification.Reason != storeforward.ReasonRetained {
			t.Fatalf("expected record retained while recipient offline, got %s", sendClassification.Reason)
		}

		recipientOnline := false
		var relayQueue []storeforward.Record
		if !recipientOnline {
			relayQueue = append(relayQueue, record)
		}

		reconnectAt := now.Add(5 * time.Minute)
		recipientOnline = true
		var delivered []storeforward.Record
		if recipientOnline {
			for _, queued := range relayQueue {
				classification := storeforward.ClassifyPurge(queued, reconnectAt, 0, 0)
				if classification.Reason == storeforward.ReasonExpired {
					continue
				}
				delivered = append(delivered, queued)
			}
		}

		if len(delivered) != 1 {
			t.Fatalf("expected exactly one offline-delivered record, got %d", len(delivered))
		}
		if string(delivered[0].Payload) != "ciphertext:hello" {
			t.Fatalf("unexpected delivered payload: %q", delivered[0].Payload)
		}

		peers := make([]storeforward.PeerInfo, 18)
		for i := range peers {
			peers[i] = storeforward.PeerInfo{ID: fmt.Sprintf("peer-%02d", i), Score: i}
		}
		plan := storeforward.BuildReplicationPlan(record, peers)
		if !plan.NeedsRepair() {
			t.Fatalf("expected repair when peers < target")
		}
	})

	t.Run("E2E-HS-01", func(t *testing.T) {
		h := historysync.NewHandler()
		req := historysync.Request{ServerID: "server-1", Epoch: 3, Cursor: "start"}
		resp := h.Handle(req)
		if resp.Token.HistoryRoot != history.CanonicalRoot(req.ServerID, req.Epoch, history.ModeEpochHistory) {
			t.Fatalf("unexpected history root")
		}
		if !resp.Locked.Segment.Locked {
			t.Fatalf("expected locked segment")
		}
		builder := merkle.NewBuilder()
		builder.AddLeaf([]byte("leaf-0"))
		builder.AddLeaf([]byte("leaf-1"))
		proof, err := builder.Proof(0)
		if err != nil {
			t.Fatalf("proof generation failed: %v", err)
		}
		if err := merkle.VerifyProof(builder.Root(), []byte("leaf-0"), proof, 0, 2); err != nil {
			t.Fatalf("proof verification failed: %v", err)
		}
	})

	t.Run("E2E-SR-01", func(t *testing.T) {
		idx := search.NewIndex()
		idx.Insert(search.Document{ID: "d1", Scope: "scope-S7-feed", Body: "query match", From: "user:alice", HasFile: true, HasLink: false, CreatedAt: time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)})
		idx.Insert(search.Document{ID: "d2", Scope: "scope-S7-feed", Body: "query match with link", From: "user:alice", HasFile: true, HasLink: true, CreatedAt: time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)})
		filters := search.QueryFilters{FromUser: "user:alice", HasFile: true, HasLink: true}
		results, err := idx.Query("scope-S7-feed", "query", filters, search.Pagination{Limit: 10})
		if err != nil {
			t.Fatalf("query failed: %v", err)
		}
		if len(results) != 1 || results[0].ID != "d2" {
			t.Fatalf("unexpected results: %v", results)
		}
		if idx.MigrationStatus() != "pending" {
			t.Fatalf("expected pending migration by default")
		}
		idx.ApplyMigration("v07-search")
		if idx.MigrationStatus() != "applied" {
			t.Fatalf("expected migration status applied")
		}
	})

	t.Run("E2E-PR-01", func(t *testing.T) {
		adapter := &push.MockAdapter{}
		relay := push.NewService(adapter)
		if err := relay.Register(context.Background(), push.Registration{DeviceID: "device", ClientID: "client", Token: "token"}); err != nil {
			t.Fatalf("registration failed: %v", err)
		}
		payload := push.BuildPayload("env-1", "device", []byte("payload"), map[string]string{"notif": "1", push.AuthMetadataKey: "token"})
		if err := relay.Forward(context.Background(), payload); err != nil {
			t.Fatalf("forward failed: %v", err)
		}
		if len(adapter.Sent) != 1 {
			t.Fatalf("expected one forwarded payload")
		}
		pipe := notification.NewPipeline(10 * time.Second)
		action, status := pipe.Fire(notification.Notification{ServerID: "srv", ChannelID: "chan", MessageID: "msg", Trigger: notification.TriggerPush}, time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC))
		if status != "notification.push.ready" {
			t.Fatalf("unexpected status %s", status)
		}
		if action.Target.ServerID != "srv" {
			t.Fatalf("unexpected action target %v", action.Target)
		}
	})

	t.Run("E2E-EP-01", func(t *testing.T) {
		h := historysync.NewHandler()
		req := historysync.Request{ServerID: "server-edge", Epoch: 5, Cursor: "edge"}
		resp := h.Handle(req)
		if !resp.Locked.Segment.Locked || resp.Locked.Reason != "keys_missing" {
			t.Fatalf("expected locked history reason with keys_missing")
		}
		if desc := historysync.ModeEpochDescription(resp.Locked.Segment); !strings.Contains(desc, "status=locked") {
			t.Fatalf("unexpected mode epoch description %s", desc)
		}
		capsule := history.NewCapsuleMetadata("server-edge", resp.Token.Epoch, history.ModeEpochHistory)
		if capsule.Root != resp.Locked.Capsule.Root {
			t.Fatalf("capsule root drift")
		}
		if resp.Locked.IsAvailable {
			t.Fatalf("locked history should not be available")
		}
		if !archivist.CanTransition(archivist.StateEnrolling, archivist.StateActive) {
			t.Fatalf("expected archivist to join")
		}
	})
}
