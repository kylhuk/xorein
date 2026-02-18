package ui

import (
	"errors"
	"testing"

	coreui "github.com/aether/code_aether/pkg/ui"
)

func TestNewIdentityFlowRequiresShell(t *testing.T) {
	t.Parallel()

	_, err := NewIdentityFlow(nil)
	if !errors.Is(err, ErrShellRequired) {
		t.Fatalf("expected ErrShellRequired, got %v", err)
	}
}

func TestIdentityFlowOnboardingAndCreate(t *testing.T) {
	t.Parallel()

	shell := coreui.NewShell()
	flow, err := NewIdentityFlow(shell)
	if err != nil {
		t.Fatalf("NewIdentityFlow returned error: %v", err)
	}

	if err := flow.StartOnboarding(); err != nil {
		t.Fatalf("StartOnboarding returned error: %v", err)
	}
	if err := flow.AcknowledgeNoPasswordResetWarning(); err != nil {
		t.Fatalf("AcknowledgeNoPasswordResetWarning returned error: %v", err)
	}
	if err := flow.CreateIdentity("alice"); err != nil {
		t.Fatalf("CreateIdentity returned error: %v", err)
	}

	state, err := flow.State()
	if err != nil {
		t.Fatalf("State returned error: %v", err)
	}
	if state.IdentityDisplay != "alice" {
		t.Fatalf("unexpected identity display: %q", state.IdentityDisplay)
	}
	if !state.NoPasswordResetWarningAck {
		t.Fatal("expected warning acknowledgement")
	}
}

func TestIdentityFlowCompleteRestore(t *testing.T) {
	t.Parallel()

	shell := coreui.NewShell()
	flow, err := NewIdentityFlow(shell)
	if err != nil {
		t.Fatalf("NewIdentityFlow returned error: %v", err)
	}

	if err := flow.CompleteRestore("restored-user", "", "backup-outdated"); err != nil {
		t.Fatalf("CompleteRestore returned error: %v", err)
	}

	state, err := flow.State()
	if err != nil {
		t.Fatalf("State returned error: %v", err)
	}
	if !state.RestoreCompleted {
		t.Fatal("expected restore completed flag")
	}
	if state.LastBackupRestoreReason != "backup-outdated" {
		t.Fatalf("unexpected backup reason: %q", state.LastBackupRestoreReason)
	}
}
