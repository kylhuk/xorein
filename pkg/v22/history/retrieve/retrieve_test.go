package retrieve

import (
	"testing"

	"github.com/aether/code_aether/pkg/v22/history/integrity"
)

func TestRetrievalAuthorized(t *testing.T) {
	store := NewRetrievalStore(5)
	store.RegisterMembership("space", "secret")

	manifest := integrity.HistorySegmentManifest{
		SpaceID:   "space",
		ChannelID: "chan",
		Segments: []integrity.HistorySegment{
			{ID: "seg-1", Hash: "hash"},
		},
	}
	manifest.RegenerateHash()

	head := integrity.HistoryHead{
		SpaceID:   "space",
		ChannelID: "chan",
		Manifest:  manifest,
	}
	head.Signature = integrity.DeriveHeadSignature(manifest.ManifestHash, "secret")

	store.StoreManifest(manifest)
	store.StoreHead(head)
	store.StoreSegment("space", "chan", "seg-1", []byte("payload"))

	req := RetrievalRequest{
		SpaceID:   "space",
		ChannelID: "chan",
		Key:       DeriveRetrievalKey("space", "secret"),
	}

	if _, err := store.RetrieveHead(req); err != nil {
		t.Fatalf("RetrieveHead failed: %v", err)
	}

	if _, err := store.RetrieveManifest(req); err != nil {
		t.Fatalf("RetrieveManifest failed: %v", err)
	}

	if payload, err := store.RetrieveSegment(req, "seg-1"); err != nil {
		t.Fatalf("RetrieveSegment failed: %v", err)
	} else if string(payload) != "payload" {
		t.Fatalf("unexpected payload %s", payload)
	}
}

func TestRetrieveUnauthorizedIsGeneric(t *testing.T) {
	store := NewRetrievalStore(5)
	store.RegisterMembership("space", "secret")

	req := RetrievalRequest{
		SpaceID:   "space",
		ChannelID: "chan",
		Key:       "wrong",
	}

	if _, err := store.RetrieveHead(req); err != ErrRetrievalFailure {
		t.Fatalf("expected generic failure, got %v", err)
	}

	if _, err := store.RetrieveManifest(req); err != ErrRetrievalFailure {
		t.Fatalf("expected generic failure, got %v", err)
	}
}

func TestRateLimiterBlocks(t *testing.T) {
	store := NewRetrievalStore(1)
	store.RegisterMembership("space", "secret")
	req := RetrievalRequest{SpaceID: "space", ChannelID: "chan", Key: DeriveRetrievalKey("space", "secret")}

	if _, err := store.RetrieveHead(req); err != ErrRetrievalFailure {
		t.Fatalf("expected head failure due to missing data, got %v", err)
	}

	if _, err := store.RetrieveHead(req); err != ErrRetrievalRateLimited {
		t.Fatalf("expected rate limited, got %v", err)
	}
}
