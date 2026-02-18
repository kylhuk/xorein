package ui

import (
	"errors"
	"testing"
)

func TestCreateIdentityWithWarningRequiresAcknowledgement(t *testing.T) {
	t.Parallel()

	shell := NewShell()
	if err := shell.BeginIdentityOnboarding(); err != nil {
		t.Fatalf("BeginIdentityOnboarding returned error: %v", err)
	}

	err := shell.CreateIdentityWithWarning("alice")
	if !errors.Is(err, ErrNoPasswordResetWarningRequired) {
		t.Fatalf("expected ErrNoPasswordResetWarningRequired, got %v", err)
	}

	if err := shell.AcknowledgeNoPasswordResetWarning(); err != nil {
		t.Fatalf("AcknowledgeNoPasswordResetWarning returned error: %v", err)
	}
	if err := shell.CreateIdentityWithWarning("alice"); err != nil {
		t.Fatalf("CreateIdentityWithWarning returned error: %v", err)
	}

	state := shell.State()
	if !state.NoPasswordResetWarningShown {
		t.Fatal("expected warning to be shown")
	}
	if !state.NoPasswordResetWarningAck {
		t.Fatal("expected warning acknowledgement to be true")
	}
	if state.IdentityDisplay != "alice" {
		t.Fatalf("unexpected identity display: %q", state.IdentityDisplay)
	}
	if state.NoPasswordResetWarningText != MandatoryNoPasswordResetWarning {
		t.Fatalf("unexpected warning text: %q", state.NoPasswordResetWarningText)
	}
}

func TestRecordRestoreOutcomeValidatesReasons(t *testing.T) {
	t.Parallel()

	shell := NewShell()

	err := shell.RecordRestoreOutcome("identity-unknown", "backup-password")
	if !errors.Is(err, ErrIdentityReasonInvalid) {
		t.Fatalf("expected ErrIdentityReasonInvalid, got %v", err)
	}

	err = shell.RecordRestoreOutcome("identity-mismatch", "backup-unknown")
	if !errors.Is(err, ErrBackupReasonInvalid) {
		t.Fatalf("expected ErrBackupReasonInvalid, got %v", err)
	}

	if err := shell.RecordRestoreOutcome("identity-mismatch", "backup-corrupt"); err != nil {
		t.Fatalf("RecordRestoreOutcome returned error: %v", err)
	}

	state := shell.State()
	if state.LastIdentityRestoreReason != "identity-mismatch" {
		t.Fatalf("unexpected identity reason: %q", state.LastIdentityRestoreReason)
	}
	if state.LastBackupRestoreReason != "backup-corrupt" {
		t.Fatalf("unexpected backup reason: %q", state.LastBackupRestoreReason)
	}
}

func TestCompleteRestoreMarksRecoveryState(t *testing.T) {
	t.Parallel()

	shell := NewShell()
	if err := shell.AcknowledgeNoPasswordResetWarning(); err != nil {
		t.Fatalf("AcknowledgeNoPasswordResetWarning returned error: %v", err)
	}

	if err := shell.CompleteRestore("restored-user", "", ""); err != nil {
		t.Fatalf("CompleteRestore returned error: %v", err)
	}

	state := shell.State()
	if !state.RestoreCompleted {
		t.Fatal("expected RestoreCompleted to be true")
	}
	if state.IdentityDisplay != "restored-user" {
		t.Fatalf("unexpected identity display: %q", state.IdentityDisplay)
	}
}
