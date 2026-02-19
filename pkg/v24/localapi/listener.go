package localapi

import (
	"fmt"
	"strings"
)

// RefusalReason is the deterministic refusal taxonomy for local API scaffolding.
type RefusalReason int

const (
	RefusalReasonUnspecified RefusalReason = iota
	RefusalReasonNonLocalBind
	RefusalReasonInvalidToken
	RefusalReasonUnauthorizedCapability
	RefusalReasonOwnershipMismatch
	RefusalReasonVersionDowngrade
	RefusalReasonNonceReplay
	RefusalReasonMalformedFrame
	RefusalReasonOversizeFrame
)

func (r RefusalReason) String() string {
	switch r {
	case RefusalReasonNonLocalBind:
		return "non-local bind"
	case RefusalReasonInvalidToken:
		return "invalid token"
	case RefusalReasonUnauthorizedCapability:
		return "unauthorized capability"
	case RefusalReasonOwnershipMismatch:
		return "ownership mismatch"
	case RefusalReasonVersionDowngrade:
		return "version downgrade"
	case RefusalReasonNonceReplay:
		return "nonce replay"
	case RefusalReasonMalformedFrame:
		return "malformed frame"
	case RefusalReasonOversizeFrame:
		return "oversize frame"
	default:
		return "unspecified"
	}
}

// RefusalError wraps deterministic refusal metadata without exposing payloads.
type RefusalError struct {
	Reason RefusalReason
	Detail string
}

func (e RefusalError) Error() string {
	return fmt.Sprintf("refused: %s (%s)", e.Reason, e.Detail)
}

// ListenerConfig configures how the daemon binds to its local transport.
type ListenerConfig struct {
	Network          string
	Address          string
	ExpectedOwnerUID *int
	ActualOwnerUID   *int
}

// ValidateLocalBind enforces that only local transports are allowed.
func (c ListenerConfig) ValidateLocalBind() error {
	network := strings.ToLower(c.Network)
	switch network {
	case "unix", "npipe":
		if err := c.ValidateOwnership(); err != nil {
			return err
		}
		return nil
	default:
		return RefusalError{
			Reason: RefusalReasonNonLocalBind,
			Detail: fmt.Sprintf("network %s is not local", c.Network),
		}
	}
}

func (c ListenerConfig) ValidateOwnership() error {
	if c.ExpectedOwnerUID == nil || c.ActualOwnerUID == nil {
		return nil
	}
	if *c.ExpectedOwnerUID != *c.ActualOwnerUID {
		return RefusalError{
			Reason: RefusalReasonOwnershipMismatch,
			Detail: fmt.Sprintf("expected uid=%d actual=%d", *c.ExpectedOwnerUID, *c.ActualOwnerUID),
		}
	}
	return nil
}
