package peer

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	proto "github.com/aether/code_aether/pkg/v0_1/protocol"
)

// relayOpacityModes are the modes that require opaque (ciphertext-only) delivery bodies.
var relayOpacityModes = map[string]bool{
	"seal":        true,
	"tree":        true,
	"crowd":       true,
	"channel":     true,
	"mediashield": true,
}

// RelayDelivery is the minimal structure needed to check relay opacity.
type RelayDelivery struct {
	Kind      string `json:"kind"`
	ScopeID   string `json:"scope_id"`
	ScopeType string `json:"scope_type"`
	Body      string `json:"body"`
	Mode      string `json:"mode,omitempty"`
}

// CheckRelayOpacity verifies that a relay Delivery for an encrypted mode
// does not contain plaintext in its body field.
// Returns CodeRelayOpacityViolation if the body appears to be plaintext.
func CheckRelayOpacity(deliveryJSON []byte) *proto.PeerStreamError {
	var d RelayDelivery
	if err := json.Unmarshal(deliveryJSON, &d); err != nil {
		// Malformed delivery — let the operation handler reject it.
		return nil
	}

	// Only check opacity for encrypted modes.
	if !relayOpacityModes[d.Mode] && d.ScopeType != "dm" && d.ScopeType != "group" {
		return nil
	}

	body := strings.TrimSpace(d.Body)
	if body == "" {
		return nil
	}

	// Body must be valid base64url without padding (ciphertext).
	decoded, err := base64.RawURLEncoding.DecodeString(body)
	if err != nil {
		// Not valid base64url → likely plaintext → violation.
		return proto.NewPeerStreamError(proto.CodeRelayOpacityViolation,
			"delivery body is not base64url-encoded ciphertext", nil)
	}

	// A ciphertext body must be at least 16 bytes (AEAD tag minimum).
	if len(decoded) < 16 {
		return proto.NewPeerStreamError(proto.CodeRelayOpacityViolation,
			"delivery body too short to be valid ciphertext", nil)
	}

	// Heuristic: reject if the decoded bytes look like printable UTF-8
	// that's longer than 64 chars — a relay sees only ciphertext, never readable text.
	if looksLikePlaintext(decoded) {
		return proto.NewPeerStreamError(proto.CodeRelayOpacityViolation,
			"delivery body appears to be plaintext", nil)
	}

	return nil
}

// looksLikePlaintext returns true if >70% of bytes are printable ASCII,
// suggesting the decoded body is plaintext rather than ciphertext.
func looksLikePlaintext(b []byte) bool {
	if len(b) < 16 {
		return false
	}
	printable := 0
	for _, c := range b {
		if c >= 0x20 && c < 0x7F {
			printable++
		}
	}
	return printable*10 > len(b)*7 // >70%
}
