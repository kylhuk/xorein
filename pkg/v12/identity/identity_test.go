package identity

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateFromSeedDeterministic(t *testing.T) {
	t.Parallel()

	seed := bytes.Repeat([]byte{0x11}, SeedSize)
	now := time.Unix(1_730_000_000, 0).UTC()

	recordA, _, err := CreateFromSeed(seed, now, "")
	if err != nil {
		t.Fatalf("CreateFromSeed returned error: %v", err)
	}
	recordB, _, err := CreateFromSeed(seed, now, "")
	if err != nil {
		t.Fatalf("CreateFromSeed returned error: %v", err)
	}

	if recordA.IdentityID != recordB.IdentityID {
		t.Fatalf("identity ids differ: %q vs %q", recordA.IdentityID, recordB.IdentityID)
	}
	if recordA.PublicKeyFingerprint != recordB.PublicKeyFingerprint {
		t.Fatalf("fingerprints differ: %q vs %q", recordA.PublicKeyFingerprint, recordB.PublicKeyFingerprint)
	}
	if recordA.KeyReference == "" {
		t.Fatal("expected default key reference")
	}
}

func TestCreateFromSeedRejectsInvalidSeed(t *testing.T) {
	t.Parallel()

	_, _, err := CreateFromSeed([]byte("short-seed"), time.Now().UTC(), "")
	if !errors.Is(err, ErrSeedInvalid) {
		t.Fatalf("expected ErrSeedInvalid, got %v", err)
	}
}

func TestEnsureImmutableCreatesAndLoads(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "identity.json")
	seed := bytes.Repeat([]byte{0x22}, SeedSize)
	now := time.Unix(1_730_000_100, 0).UTC()

	created, err := EnsureImmutable(path, seed, now, "file://keyref")
	if err != nil {
		t.Fatalf("EnsureImmutable(create) returned error: %v", err)
	}

	loaded, err := EnsureImmutable(path, seed, now.Add(5*time.Minute), "file://different-keyref")
	if err != nil {
		t.Fatalf("EnsureImmutable(load) returned error: %v", err)
	}

	if created.IdentityID != loaded.IdentityID {
		t.Fatalf("expected same identity id, got %q vs %q", created.IdentityID, loaded.IdentityID)
	}
	if created.PublicKeyFingerprint != loaded.PublicKeyFingerprint {
		t.Fatalf("expected same fingerprint, got %q vs %q", created.PublicKeyFingerprint, loaded.PublicKeyFingerprint)
	}
}

func TestEnsureImmutableRejectsDuplicateSeed(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "identity.json")
	seedA := bytes.Repeat([]byte{0x33}, SeedSize)
	seedB := bytes.Repeat([]byte{0x44}, SeedSize)
	now := time.Unix(1_730_000_200, 0).UTC()

	if _, err := EnsureImmutable(path, seedA, now, ""); err != nil {
		t.Fatalf("EnsureImmutable(create) returned error: %v", err)
	}

	_, err := EnsureImmutable(path, seedB, now, "")
	if !errors.Is(err, ErrIdentityDuplicate) {
		t.Fatalf("expected ErrIdentityDuplicate, got %v", err)
	}
}

func TestLoadRejectsCorruptState(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "identity.json")
	if err := os.WriteFile(path, []byte("{not-json"), 0o600); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	_, err := Load(path)
	if !errors.Is(err, ErrIdentityCorrupt) {
		t.Fatalf("expected ErrIdentityCorrupt, got %v", err)
	}
}
