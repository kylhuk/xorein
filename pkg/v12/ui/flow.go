package ui

import (
	"errors"
	"strings"

	coreui "github.com/aether/code_aether/pkg/ui"
)

var ErrShellRequired = errors.New("shell required")

// IdentityFlow wraps shell onboarding/recovery state transitions for v12.
type IdentityFlow struct {
	shell *coreui.Shell
}

// NewIdentityFlow binds a flow to an existing shell state instance.
func NewIdentityFlow(shell *coreui.Shell) (*IdentityFlow, error) {
	if shell == nil {
		return nil, ErrShellRequired
	}
	return &IdentityFlow{shell: shell}, nil
}

// StartOnboarding enters identity onboarding and shows the no-reset warning.
func (f *IdentityFlow) StartOnboarding() error {
	if f == nil || f.shell == nil {
		return ErrShellRequired
	}
	return f.shell.BeginIdentityOnboarding()
}

// AcknowledgeNoPasswordResetWarning records explicit warning acknowledgement.
func (f *IdentityFlow) AcknowledgeNoPasswordResetWarning() error {
	if f == nil || f.shell == nil {
		return ErrShellRequired
	}
	return f.shell.AcknowledgeNoPasswordResetWarning()
}

// CreateIdentity enforces warning acknowledgement before identity creation.
func (f *IdentityFlow) CreateIdentity(display string) error {
	if f == nil || f.shell == nil {
		return ErrShellRequired
	}
	return f.shell.CreateIdentityWithWarning(display)
}

// CompleteRestore applies deterministic reason labels and marks restore complete.
func (f *IdentityFlow) CompleteRestore(display, identityReason, backupReason string) error {
	if f == nil || f.shell == nil {
		return ErrShellRequired
	}
	return f.shell.CompleteRestore(display, strings.TrimSpace(identityReason), strings.TrimSpace(backupReason))
}

// State returns the current shell state snapshot.
func (f *IdentityFlow) State() (coreui.AppState, error) {
	if f == nil || f.shell == nil {
		return coreui.AppState{}, ErrShellRequired
	}
	return f.shell.State(), nil
}
