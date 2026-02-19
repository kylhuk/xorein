package integrity

import (
	"errors"
	"testing"
	"time"
)

func baseManifest() Manifest {
	manifest := Manifest{
		SpaceID:   "prod-space",
		ChannelID: "chan-A",
		Segments: []ManifestSegment{
			{
				ID:    "seg-a",
				Start: time.Unix(0, 0).UTC(),
				End:   time.Unix(3600, 0).UTC(),
				Hash:  "hash-a",
			},
		},
	}
	manifest.RegenerateHash()
	return manifest
}

func requireManifestCode(t *testing.T, err error, want ManifestValidationCode) {
	t.Helper()
	var ve *ManifestValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ManifestValidationError, got %v", err)
	}
	if ve.Code != want {
		t.Fatalf("expected code %s, got %s", want, ve.Code)
	}
}

func requireHeadCode(t *testing.T, err error, want HeadValidationCode) {
	t.Helper()
	var he *HeadValidationError
	if !errors.As(err, &he) {
		t.Fatalf("expected HeadValidationError, got %v", err)
	}
	if he.Code != want {
		t.Fatalf("expected code %s, got %s", want, he.Code)
	}
}

func TestManifestValidateSuccess(t *testing.T) {
	manifest := baseManifest()
	if err := manifest.Validate(); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestManifestValidateMissingSpaceID(t *testing.T) {
	manifest := baseManifest()
	manifest.SpaceID = ""
	if err := manifest.Validate(); err == nil {
		t.Fatalf("expected error for missing space id")
	} else {
		requireManifestCode(t, err, ManifestValidationCodeMissingSpaceID)
	}
}

func TestManifestValidateMissingChannelID(t *testing.T) {
	manifest := baseManifest()
	manifest.ChannelID = ""
	if err := manifest.Validate(); err == nil {
		t.Fatalf("expected error for missing channel id")
	} else {
		requireManifestCode(t, err, ManifestValidationCodeMissingChannelID)
	}
}

func TestManifestValidateSegmentHashMissing(t *testing.T) {
	manifest := baseManifest()
	manifest.Segments[0].Hash = ""
	manifest.RegenerateHash()
	if err := manifest.Validate(); err == nil {
		t.Fatalf("expected segment hash error")
	} else {
		requireManifestCode(t, err, ManifestValidationCodeSegmentHashMissing)
	}
}

func TestManifestValidateSegmentTimeOrder(t *testing.T) {
	manifest := baseManifest()
	manifest.Segments[0].Start = manifest.Segments[0].End
	manifest.RegenerateHash()
	if err := manifest.Validate(); err == nil {
		t.Fatalf("expected segment time order error")
	} else {
		requireManifestCode(t, err, ManifestValidationCodeSegmentTimeOrder)
	}
}

func TestManifestValidateHashMismatch(t *testing.T) {
	manifest := baseManifest()
	manifest.ManifestHash = "invalid"
	if err := manifest.Validate(); err == nil {
		t.Fatalf("expected manifest hash mismatch")
	} else {
		requireManifestCode(t, err, ManifestValidationCodeHashMismatch)
	}
}

func TestHeadValidateSuccess(t *testing.T) {
	manifest := baseManifest()
	key := "member-key"
	head := Head{
		SpaceID:   manifest.SpaceID,
		ChannelID: manifest.ChannelID,
		Manifest:  manifest,
		Signature: DeriveHeadSignature(manifest.ManifestHash, key),
	}
	if err := head.Validate(key); err != nil {
		t.Fatalf("expected head success, got %v", err)
	}
}

func TestHeadValidateMissingMembershipKey(t *testing.T) {
	manifest := baseManifest()
	head := Head{SpaceID: manifest.SpaceID, ChannelID: manifest.ChannelID, Manifest: manifest, Signature: "sig"}
	if err := head.Validate(""); err == nil {
		t.Fatalf("expected membership key missing")
	} else {
		requireHeadCode(t, err, HeadValidationCodeMissingMembershipKey)
	}
}

func TestHeadValidateMissingSignature(t *testing.T) {
	manifest := baseManifest()
	head := Head{SpaceID: manifest.SpaceID, ChannelID: manifest.ChannelID, Manifest: manifest}
	if err := head.Validate("member"); err == nil {
		t.Fatalf("expected missing signature error")
	} else {
		requireHeadCode(t, err, HeadValidationCodeMissingSignature)
	}
}

func TestHeadValidateSignatureMismatch(t *testing.T) {
	manifest := baseManifest()
	head := Head{
		SpaceID:   manifest.SpaceID,
		ChannelID: manifest.ChannelID,
		Manifest:  manifest,
		Signature: "bad-signature",
	}
	if err := head.Validate("member-key"); err == nil {
		t.Fatalf("expected signature mismatch")
	} else {
		requireHeadCode(t, err, HeadValidationCodeSignatureMismatch)
	}
}

func TestHeadValidateSpaceMismatch(t *testing.T) {
	manifest := baseManifest()
	key := "member-key"
	head := Head{
		SpaceID:   "other",
		ChannelID: manifest.ChannelID,
		Manifest:  manifest,
		Signature: DeriveHeadSignature(manifest.ManifestHash, key),
	}
	if err := head.Validate(key); err == nil {
		t.Fatalf("expected space mismatch")
	} else {
		requireHeadCode(t, err, HeadValidationCodeSpaceMismatch)
	}
}

func TestHeadValidateChannelMismatch(t *testing.T) {
	manifest := baseManifest()
	key := "member-key"
	head := Head{
		SpaceID:   manifest.SpaceID,
		ChannelID: "other-channel",
		Manifest:  manifest,
		Signature: DeriveHeadSignature(manifest.ManifestHash, key),
	}
	if err := head.Validate(key); err == nil {
		t.Fatalf("expected channel mismatch")
	} else {
		requireHeadCode(t, err, HeadValidationCodeChannelMismatch)
	}
}

func TestHeadValidateManifestInvalid(t *testing.T) {
	manifest := baseManifest()
	manifest.Segments = nil
	manifest.ManifestHash = ""
	head := Head{
		SpaceID:   manifest.SpaceID,
		ChannelID: manifest.ChannelID,
		Manifest:  manifest,
		Signature: "sig",
	}
	if err := head.Validate("member-key"); err == nil {
		t.Fatalf("expected invalid manifest")
	} else {
		if errr := errors.Unwrap(err); errr == nil {
			t.Fatalf("expected unwrap to succeed")
		}
		requireHeadCode(t, err, HeadValidationCodeInvalidManifest)
	}
}
