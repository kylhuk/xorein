package phase6

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	deeplinkServerIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,64}$`)
)

// DeeplinkValidationError describes why a supplied deeplink failed validation.
type DeeplinkValidationError struct {
	Reason string
}

func (e *DeeplinkValidationError) Error() string {
	return fmt.Sprintf("deeplink validation: %s", e.Reason)
}

// DeepLink represents the parsed outcome of a join deeplink.
type DeepLink struct {
	ServerID string
}

// ParseJoinDeepLink validates the expected aether join format (aether://join/<server-id>) and
// returns the parsed server identifier.
func ParseJoinDeepLink(raw string) (*DeepLink, error) {
	if raw == "" {
		return nil, &DeeplinkValidationError{Reason: "empty deeplink"}
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid deeplink: %w", err)
	}

	if strings.ToLower(u.Scheme) != "aether" {
		return nil, &DeeplinkValidationError{Reason: "invalid scheme, expected aether"}
	}

	if strings.ToLower(u.Host) != "join" {
		return nil, &DeeplinkValidationError{Reason: "deeplink host must be join"}
	}

	trimmed := strings.Trim(u.Path, "/")
	if trimmed == "" {
		return nil, &DeeplinkValidationError{Reason: "missing server identifier"}
	}

	if !deeplinkServerIDRegex.MatchString(trimmed) {
		return nil, &DeeplinkValidationError{Reason: "server identifier invalid (alphanumeric/_/- only, 3-64 chars)"}
	}

	return &DeepLink{ServerID: trimmed}, nil
}
