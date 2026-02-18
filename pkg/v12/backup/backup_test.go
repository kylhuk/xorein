package backup

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v12/identity"
)

func TestExportRestoreRoundTrip(t *testing.T) {
	t.Parallel()

	payload := testPayload(t, 0x11)
	now := time.Unix(1_730_001_000, 0).UTC()

	envelope, err := Export(payload, "backup-password", now)
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}
	if !strings.HasPrefix(envelope.BackupID, "BKP-") {
		t.Fatalf("expected backup id prefix BKP-, got %q", envelope.BackupID)
	}

	data, err := MarshalEnvelope(envelope)
	if err != nil {
		t.Fatalf("MarshalEnvelope returned error: %v", err)
	}

	restored, err := Restore(data, "backup-password", nil)
	if err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}
	if restored.Identity.IdentityID != payload.Identity.IdentityID {
		t.Fatalf("identity mismatch: got %q want %q", restored.Identity.IdentityID, payload.Identity.IdentityID)
	}
	if restored.Config["theme"] != "amber" {
		t.Fatalf("config mismatch: got %q", restored.Config["theme"])
	}
}

func TestRestoreWrongPasswordReturnsDeterministicReason(t *testing.T) {
	t.Parallel()

	payload := testPayload(t, 0x22)
	envelope, err := Export(payload, "correct-password", time.Unix(1_730_001_100, 0).UTC())
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}
	data, err := MarshalEnvelope(envelope)
	if err != nil {
		t.Fatalf("MarshalEnvelope returned error: %v", err)
	}

	_, err = Restore(data, "wrong-password", nil)
	if !errors.Is(err, ErrBackupPassword) {
		t.Fatalf("expected ErrBackupPassword, got %v", err)
	}
	if reason := ReasonFromError(err); reason != ReasonBackupPassword {
		t.Fatalf("unexpected reason: got %q", reason)
	}
}

func TestRestoreDetectsTamperedCiphertext(t *testing.T) {
	t.Parallel()

	payload := testPayload(t, 0x33)
	envelope, err := Export(payload, "backup-password", time.Unix(1_730_001_200, 0).UTC())
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}
	envelope.CiphertextSHA256 = strings.Repeat("0", len(envelope.CiphertextSHA256))

	data, err := MarshalEnvelope(envelope)
	if err != nil {
		t.Fatalf("MarshalEnvelope returned error: %v", err)
	}

	_, err = Restore(data, "backup-password", nil)
	if !errors.Is(err, ErrBackupCorrupt) {
		t.Fatalf("expected ErrBackupCorrupt, got %v", err)
	}
	if reason := ReasonFromError(err); reason != ReasonBackupCorrupt {
		t.Fatalf("unexpected reason: got %q", reason)
	}
}

func TestRestoreRejectsTruncatedBackup(t *testing.T) {
	t.Parallel()

	payload := testPayload(t, 0x44)
	envelope, err := Export(payload, "backup-password", time.Unix(1_730_001_300, 0).UTC())
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}
	data, err := MarshalEnvelope(envelope)
	if err != nil {
		t.Fatalf("MarshalEnvelope returned error: %v", err)
	}

	truncated := data[:len(data)/2]
	_, err = Restore(truncated, "backup-password", nil)
	if !errors.Is(err, ErrBackupCorrupt) {
		t.Fatalf("expected ErrBackupCorrupt, got %v", err)
	}
}

func TestRestoreRejectsIdentityMismatch(t *testing.T) {
	t.Parallel()

	payload := testPayload(t, 0x55)
	envelope, err := Export(payload, "backup-password", time.Unix(1_730_001_400, 0).UTC())
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}
	data, err := MarshalEnvelope(envelope)
	if err != nil {
		t.Fatalf("MarshalEnvelope returned error: %v", err)
	}

	otherIdentity := mustIdentityRecord(t, 0x66)
	_, err = Restore(data, "backup-password", &otherIdentity)
	if !errors.Is(err, ErrIdentityMismatch) {
		t.Fatalf("expected ErrIdentityMismatch, got %v", err)
	}
	if reason := ReasonFromError(err); reason != ReasonIdentityMismatch {
		t.Fatalf("unexpected reason: got %q", reason)
	}
}

func TestRestoreRejectsOutdatedVersion(t *testing.T) {
	t.Parallel()

	payload := testPayload(t, 0x77)
	envelope, err := Export(payload, "backup-password", time.Unix(1_730_001_500, 0).UTC())
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}
	envelope.Version = 0

	data, err := MarshalEnvelope(envelope)
	if err != nil {
		t.Fatalf("MarshalEnvelope returned error: %v", err)
	}

	_, err = Restore(data, "backup-password", nil)
	if !errors.Is(err, ErrBackupOutdated) {
		t.Fatalf("expected ErrBackupOutdated, got %v", err)
	}
	if reason := ReasonFromError(err); reason != ReasonBackupOutdated {
		t.Fatalf("unexpected reason: got %q", reason)
	}
}

func testPayload(t *testing.T, seedByte byte) Payload {
	t.Helper()
	record := mustIdentityRecord(t, seedByte)
	return Payload{
		Identity: record,
		Config: map[string]string{
			"locale": "en-US",
			"theme":  "amber",
		},
	}
}

func mustIdentityRecord(t *testing.T, seedByte byte) identity.Record {
	t.Helper()
	seed := bytes.Repeat([]byte{seedByte}, identity.SeedSize)
	record, _, err := identity.CreateFromSeed(seed, time.Unix(1_730_001_000, 0).UTC(), "keyring://local/identity")
	if err != nil {
		t.Fatalf("CreateFromSeed returned error: %v", err)
	}
	return record
}
