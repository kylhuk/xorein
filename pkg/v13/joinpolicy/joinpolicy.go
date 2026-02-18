package joinpolicy

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Mode defines the deterministic join contract for a Space.
type Mode string

const (
	ModeInviteOnly    Mode = "invite-only"
	ModeRequestToJoin Mode = "request-to-join"
	ModeOpen          Mode = "open"
)

var (
	ErrUnknownMode        = errors.New("unknown join policy mode")
	ErrInviteTokenMissing = errors.New("invite token required for invite-only mode")
)

// Default returns the baseline join policy for new spaces.
func Default() Mode {
	return ModeInviteOnly
}

// ParseMode normalizes a raw policy name into a Mode instance.
func ParseMode(raw string) (Mode, error) {
	normalized := Mode(strings.ToLower(strings.TrimSpace(raw)))
	if normalized == "" {
		return Default(), nil
	}

	switch normalized {
	case ModeInviteOnly, ModeRequestToJoin, ModeOpen:
		return normalized, nil
	}

	variants := []string{string(ModeInviteOnly), string(ModeOpen), string(ModeRequestToJoin)}
	sort.Strings(variants)
	return "", fmt.Errorf("%w: %q; valid modes: %s", ErrUnknownMode, raw, strings.Join(variants, ", "))
}

// RequiresInviteToken reports whether the policy enforces invite tokens.
func RequiresInviteToken(mode Mode) bool {
	return mode == ModeInviteOnly
}

// ValidateRequest enforces guardrails for join requests.
func ValidateRequest(mode Mode, requester string, token string) error {
	if strings.TrimSpace(requester) == "" {
		return fmt.Errorf("requester required")
	}
	if RequiresInviteToken(mode) {
		if strings.TrimSpace(token) == "" {
			return ErrInviteTokenMissing
		}
	}
	return nil
}

// AllowedModes returns an immutable list of supported policies.
func AllowedModes() []Mode {
	modes := []Mode{ModeInviteOnly, ModeOpen, ModeRequestToJoin}
	result := make([]Mode, len(modes))
	copy(result, modes)
	return result
}
