package v23

import (
	"errors"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v23/integrity"
)

func newManifestForTest() integrity.Manifest {
	manifest := integrity.Manifest{
		SpaceID:   "manifest-space",
		ChannelID: "manifest-channel",
		Segments: []integrity.ManifestSegment{
			{
				ID:    "segment-1",
				Start: time.Unix(0, 0).UTC(),
				End:   time.Unix(60, 0).UTC(),
				Hash:  "segment-hash",
			},
		},
	}
	manifest.RegenerateHash()
	return manifest
}

func requireManifestCode(t *testing.T, err error, want integrity.ManifestValidationCode) {
	t.Helper()
	var ve *integrity.ManifestValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ManifestValidationError, got %v", err)
	}
	if ve.Code != want {
		t.Fatalf("unexpected code %s", ve.Code)
	}
}

func TestIntegrityManifestRejectsSegmentHashMissing(t *testing.T) {
	manifest := newManifestForTest()
	manifest.Segments[0].Hash = ""
	manifest.RegenerateHash()
	if err := manifest.Validate(); err == nil {
		t.Fatal("expected validation to fail")
	} else {
		requireManifestCode(t, err, integrity.ManifestValidationCodeSegmentHashMissing)
	}
}

func TestIntegrityManifestRejectsHashMismatch(t *testing.T) {
	manifest := newManifestForTest()
	manifest.ManifestHash = "bad"
	if err := manifest.Validate(); err == nil {
		t.Fatal("expected validation to fail")
	} else {
		requireManifestCode(t, err, integrity.ManifestValidationCodeHashMismatch)
	}
}
