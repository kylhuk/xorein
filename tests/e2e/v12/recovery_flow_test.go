package v12e2e

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/aether/code_aether/pkg/v11/relaypolicy"
	"github.com/aether/code_aether/pkg/v12/backup"
	"github.com/aether/code_aether/pkg/v12/identity"
)

func TestNewDeviceRestoreScenario(t *testing.T) {
	t.Parallel()

	seed := bytes.Repeat([]byte{0x91}, identity.SeedSize)
	record, _, err := identity.CreateFromSeed(seed, time.Unix(1_730_003_000, 0).UTC(), "keyring://local/identity")
	if err != nil {
		t.Fatalf("CreateFromSeed returned error: %v", err)
	}

	envelope, err := backup.Export(backup.Payload{
		Identity: record,
		Config: map[string]string{
			"locale": "en-US",
			"theme":  "amber",
		},
	}, "backup-password", time.Unix(1_730_003_100, 0).UTC())
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}

	envelopeData, err := backup.MarshalEnvelope(envelope)
	if err != nil {
		t.Fatalf("MarshalEnvelope returned error: %v", err)
	}

	restored, err := backup.Restore(envelopeData, "backup-password", nil)
	if err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}
	if restored.Identity.IdentityID != record.IdentityID {
		t.Fatalf("unexpected restored identity: got %q want %q", restored.Identity.IdentityID, record.IdentityID)
	}
}

func TestLostPasswordWithoutBackupScenario(t *testing.T) {
	t.Parallel()

	seed := bytes.Repeat([]byte{0xA1}, identity.SeedSize)
	record, _, err := identity.CreateFromSeed(seed, time.Unix(1_730_003_200, 0).UTC(), "keyring://local/identity")
	if err != nil {
		t.Fatalf("CreateFromSeed returned error: %v", err)
	}

	envelope, err := backup.Export(backup.Payload{Identity: record}, "correct-password", time.Unix(1_730_003_300, 0).UTC())
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}
	envelopeData, err := backup.MarshalEnvelope(envelope)
	if err != nil {
		t.Fatalf("MarshalEnvelope returned error: %v", err)
	}

	_, err = backup.Restore(envelopeData, "wrong-password", nil)
	if !errors.Is(err, backup.ErrBackupPassword) {
		t.Fatalf("expected ErrBackupPassword, got %v", err)
	}
	if reason := backup.ReasonFromError(err); reason != backup.ReasonBackupPassword {
		t.Fatalf("unexpected reason: got %q", reason)
	}
}

func TestRelayBoundaryRegressionScenario(t *testing.T) {
	t.Parallel()

	allowedMode, err := relaypolicy.ParsePersistenceMode("session-metadata")
	if err != nil {
		t.Fatalf("ParsePersistenceMode(session-metadata) returned error: %v", err)
	}
	if err := relaypolicy.ValidateMode(allowedMode); err != nil {
		t.Fatalf("ValidateMode(session-metadata) returned error: %v", err)
	}

	forbiddenMode, err := relaypolicy.ParsePersistenceMode("durable-message-body")
	if err != nil {
		t.Fatalf("ParsePersistenceMode(durable-message-body) returned error: %v", err)
	}
	err = relaypolicy.ValidateMode(forbiddenMode)
	var validationErr *relaypolicy.ValidationError
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if len(validationErr.ForbiddenClasses) == 0 {
		t.Fatal("expected forbidden classes in validation error")
	}
}
