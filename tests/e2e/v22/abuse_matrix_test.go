package v22

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v22/archivist/replicate"
	"github.com/aether/code_aether/pkg/v22/archivist/store"
	"github.com/aether/code_aether/pkg/v22/history/backfill"
	"github.com/aether/code_aether/pkg/v22/history/integrity"
	"github.com/aether/code_aether/pkg/v22/history/retrieve"
)

func TestPrivateSpaceAntiEnumeration(t *testing.T) {
	space := "private-space"
	channel := "chan"
	secret := "membership-secret"
	store := retrieve.NewRetrievalStore(2)
	store.RegisterMembership(space, secret)

	manifest := integrity.HistorySegmentManifest{
		SpaceID:   space,
		ChannelID: channel,
		Segments: []integrity.HistorySegment{{
			ID:    "seg-1",
			Hash:  "hash-1",
			Start: time.Unix(0, 0),
			End:   time.Unix(1, 0),
		}},
	}
	manifest.RegenerateHash()
	head := integrity.HistoryHead{
		SpaceID:   space,
		ChannelID: channel,
		Manifest:  manifest,
		Signature: integrity.DeriveHeadSignature(manifest.ManifestHash, secret),
	}
	store.StoreManifest(manifest)
	store.StoreHead(head)

	if _, err := store.RetrieveHead(retrieve.RetrievalRequest{SpaceID: space, ChannelID: channel, Key: "wrong"}); err != retrieve.ErrRetrievalFailure {
		t.Fatalf("expected anti-enumeration failure, got %v", err)
	}

	ghostKey := retrieve.DeriveRetrievalKey("ghost", "secret")
	if _, err := store.RetrieveManifest(retrieve.RetrievalRequest{SpaceID: "ghost", ChannelID: channel, Key: ghostKey}); err != retrieve.ErrRetrievalFailure {
		t.Fatalf("unexpected error for ghost space: %v", err)
	}

	validKey := retrieve.DeriveRetrievalKey(space, secret)
	retrieved, err := store.RetrieveHead(retrieve.RetrievalRequest{SpaceID: space, ChannelID: channel, Key: validKey})
	if err != nil {
		t.Fatalf("expected head retrieval to succeed, got %v", err)
	}
	if retrieved.SpaceID != space {
		t.Fatalf("unexpected space id %s", retrieved.SpaceID)
	}
}

func TestManifestAndHeadSignatureForgery(t *testing.T) {
	space := "manifest-space"
	channel := "manifest-chan"
	secret := "signing-key"
	manifest := integrity.HistorySegmentManifest{
		SpaceID:   space,
		ChannelID: channel,
		Segments: []integrity.HistorySegment{{
			ID:    "seg",
			Hash:  "hash",
			Start: time.Unix(0, 0),
			End:   time.Unix(1, 0),
		}},
	}
	manifest.RegenerateHash()
	head := integrity.HistoryHead{
		SpaceID:   space,
		ChannelID: channel,
		Manifest:  manifest,
		Signature: integrity.DeriveHeadSignature(manifest.ManifestHash, secret),
	}
	if err := head.VerifySignature(secret); err != nil {
		t.Fatalf("expected signature verification to pass, got %v", err)
	}

	forgedManifest := head
	forgedManifest.Manifest.ManifestHash = "bad"
	if err := forgedManifest.VerifySignature(secret); !errors.Is(err, integrity.ErrManifestHashMismatch) {
		t.Fatalf("expected manifest mismatch, got %v", err)
	}

	invalidSignature := head
	invalidSignature.Signature = "nop"
	if err := invalidSignature.VerifySignature(secret); !errors.Is(err, integrity.ErrHistoryHeadInvalidSignature) {
		t.Fatalf("expected invalid signature error, got %v", err)
	}
}

func TestQuotaRetentionEnforcementAbuse(t *testing.T) {
	base := time.Unix(0, 0)
	s := store.NewStore(store.Config{
		QuotaPerSpace: map[store.SpaceID]int64{"space": 10},
		ChannelCap:    map[store.ChannelID]int64{"chan": 10},
		Retention:     time.Minute,
		Now:           func() time.Time { return base },
	})
	unhandled := s.Put("space", "chan", "seg-a", 12)
	if unhandled == nil {
		t.Fatalf("expected quota/segment rejection")
	}
	if se, ok := unhandled.(store.StoreError); !ok || se.Reason != store.ReasonQuotaExceeded {
		t.Fatalf("unexpected quota rejection %v", unhandled)
	}
	if err := s.Put("space", "chan", "seg-b", 5); err != nil {
		t.Fatalf("unexpected put error %v", err)
	}
	if err := s.Put("space", "chan", "seg-c", 5); err != nil {
		t.Fatalf("unexpected put error %v", err)
	}
	base = base.Add(2 * time.Minute)
	pruned := s.Prune()
	if len(pruned) != 2 {
		t.Fatalf("expected pruning to evict two segments, got %d", len(pruned))
	}
	for _, entry := range pruned {
		if entry.Reason != store.ReasonRetentionPolicy {
			t.Fatalf("unexpected prune reason %s", entry.Reason)
		}
	}
}

func TestBackfillRequestsTimeRangeOnly(t *testing.T) {
	reqType := reflect.TypeOf(backfill.BackfillRequest{})
	allowed := map[string]struct{}{
		"SpaceID":     {},
		"ChannelID":   {},
		"Range":       {},
		"MaxSegments": {},
	}
	for i := 0; i < reqType.NumField(); i++ {
		field := reqType.Field(i)
		if _, ok := allowed[field.Name]; !ok {
			t.Fatalf("unexpected backfill request field %s", field.Name)
		}
	}
	rangeType := reflect.TypeOf(backfill.TimeRange{})
	for i := 0; i < rangeType.NumField(); i++ {
		field := rangeType.Field(i)
		if field.Name != "Start" && field.Name != "End" {
			t.Fatalf("unexpected time range field %s", field.Name)
		}
	}
}

func TestReplicaDegradedHealingSignals(t *testing.T) {
	policy := replicate.Policy{R: 3, RMin: 2}
	result := replicate.Replicate(policy, []replicate.EndpointID{"a", "b"}, func(replicate.EndpointID) error {
		return nil
	})
	if result.Health != replicate.HealthDegraded || result.Reason != replicate.ResultReplicaWritePartial || result.TargetMet {
		t.Fatalf("unexpected degraded result %+v", result)
	}
	healing := replicate.Heal(policy, []replicate.EndpointID{"a"}, []replicate.EndpointID{"b"}, func(replicate.EndpointID) error {
		return nil
	})
	if len(healing.HealedTokens) == 0 || healing.Reason != replicate.ResultReplicaHealingInProgress {
		t.Fatalf("unexpected healing result %+v", healing)
	}
}
