package integrity

import (
	"testing"
	"time"
)

func TestManifestHashVerification(t *testing.T) {
	manifest := HistorySegmentManifest{
		SpaceID:   "space",
		ChannelID: "channel",
		Segments: []HistorySegment{
			{
				ID:    "seg-a",
				Start: time.Unix(0, 0).UTC(),
				End:   time.Unix(1, 0).UTC(),
				Hash:  "hash-a",
			},
		},
	}

	manifest.RegenerateHash()

	if err := manifest.VerifyHash(); err != nil {
		t.Fatalf("expected manifest verification to pass, got %v", err)
	}

	manifest.ManifestHash = "invalid"
	if err := manifest.VerifyHash(); err != ErrManifestHashMismatch {
		t.Fatalf("expected ErrManifestHashMismatch, got %v", err)
	}
}

func TestSegmentByIDNotFound(t *testing.T) {
	manifest := HistorySegmentManifest{
		Segments: []HistorySegment{{ID: "seg-a"}},
	}

	if _, err := manifest.SegmentByID("missing"); err != ErrSegmentNotFound {
		t.Fatalf("expected ErrSegmentNotFound, got %v", err)
	}
}

func TestHistoryHeadSignatureVerification(t *testing.T) {
	manifest := HistorySegmentManifest{
		SpaceID: "space",
		Segments: []HistorySegment{
			{ID: "seg-a", Hash: "hash"},
		},
	}
	manifest.RegenerateHash()

	head := HistoryHead{
		SpaceID:  "space",
		Manifest: manifest,
	}
	key := "secret"
	head.Signature = DeriveHeadSignature(manifest.ManifestHash, key)

	if err := head.VerifySignature(key); err != nil {
		t.Fatalf("expected valid signature, got %v", err)
	}

	if err := head.VerifySignature("wrong"); err != ErrHistoryHeadInvalidSignature {
		t.Fatalf("expected invalid signature error, got %v", err)
	}
}
