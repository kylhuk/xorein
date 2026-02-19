package v23

import (
	"errors"
	"testing"

	"github.com/aether/code_aether/pkg/v23/integrity"
)

func newHeadForTest() (integrity.Head, string) {
	manifest := newManifestForTest()
	membershipKey := "ingest-key"
	head := integrity.Head{
		SpaceID:   manifest.SpaceID,
		ChannelID: manifest.ChannelID,
		Manifest:  manifest,
		Signature: integrity.DeriveHeadSignature(manifest.ManifestHash, membershipKey),
	}
	return head, membershipKey
}

func requireHeadCode(t *testing.T, err error, want integrity.HeadValidationCode) {
	t.Helper()
	var he *integrity.HeadValidationError
	if !errors.As(err, &he) {
		t.Fatalf("expected HeadValidationError, got %v", err)
	}
	if he.Code != want {
		t.Fatalf("expected %s, got %s", want, he.Code)
	}
}

func TestIntegrityHeadRejectsSignatureMismatch(t *testing.T) {
	head, key := newHeadForTest()
	head.Signature = "mismatch"
	if err := head.Validate(key); err == nil {
		t.Fatal("expected signature mismatch")
	} else {
		requireHeadCode(t, err, integrity.HeadValidationCodeSignatureMismatch)
	}
}

func TestIntegrityHeadRejectsManifestInvalid(t *testing.T) {
	head, key := newHeadForTest()
	head.Manifest.Segments = nil
	head.Manifest.ManifestHash = ""
	if err := head.Validate(key); err == nil {
		t.Fatal("expected manifest error")
	} else {
		requireHeadCode(t, err, integrity.HeadValidationCodeInvalidManifest)
	}
}
