// Package clear implements downgrade prevention for Seal sessions per spec 04 §3.4.
package clear

import (
	"errors"
	"fmt"
)

// SecurityMode identifies a conversation security mode.
type SecurityMode string

const (
	ModeSeal        SecurityMode = "seal"
	ModeTree        SecurityMode = "tree"
	ModeCrowd       SecurityMode = "crowd"
	ModeChannel     SecurityMode = "channel"
	ModeMediaShield SecurityMode = "mediashield"
	ModeClear       SecurityMode = "clear"
)

// ErrModeIncompatible is returned when a downgrade is attempted.
var ErrModeIncompatible = errors.New("seal: MODE_INCOMPATIBLE: mode downgrade rejected")

// EnforceModeContinuity returns ErrModeIncompatible if the requested mode
// would downgrade an existing scope that is already in a stronger mode.
// Per spec 04 §3.4: once a scope is in Seal/Tree/Crowd/Channel/MediaShield,
// it must stay in that mode. Downgrade to Clear is always rejected.
func EnforceModeContinuity(existing, requested SecurityMode) error {
	if existing == "" || existing == requested {
		return nil
	}
	// Downgrade to clear is always rejected for any encrypted mode.
	if existing != ModeClear && requested == ModeClear {
		return fmt.Errorf("%w: scope is in %q mode, cannot downgrade to clear", ErrModeIncompatible, existing)
	}
	// Any mode change for an existing Seal session is rejected.
	if existing == ModeSeal && requested != ModeSeal {
		return fmt.Errorf("%w: scope is in seal mode, cannot change to %q", ErrModeIncompatible, requested)
	}
	return nil
}
