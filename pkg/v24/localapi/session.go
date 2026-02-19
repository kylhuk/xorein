package localapi

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SessionToken carries a short-lived identifier for authenticated RPCs.
type SessionToken struct {
	Value     string
	ExpiresAt time.Time
}

func (t SessionToken) Valid() bool {
	return time.Now().Before(t.ExpiresAt)
}

// HandshakeMiddleware performs deterministic client handshake handling.
type HandshakeMiddleware struct {
	version    string
	minVersion string
	mu         sync.Mutex
	tokens     map[string]SessionToken
	nonces     map[string]struct{}
}

// NewHandshakeMiddleware builds middleware for a single negotiated version.
func NewHandshakeMiddleware(version string) *HandshakeMiddleware {
	return NewHandshakeMiddlewareWithMinVersion(version, version)
}

// NewHandshakeMiddlewareWithMinVersion builds middleware with an explicit minimum version.
func NewHandshakeMiddlewareWithMinVersion(version, minVersion string) *HandshakeMiddleware {
	if minVersion == "" {
		minVersion = version
	}
	return &HandshakeMiddleware{
		version:    version,
		minVersion: minVersion,
		tokens:     map[string]SessionToken{},
		nonces:     map[string]struct{}{},
	}
}

// EstablishSession accepts a ClientHello-style version + nonce and returns a token.
func (h *HandshakeMiddleware) EstablishSession(clientVersion, clientNonce string) (SessionToken, error) {
	if err := h.ensureSupportedVersion(clientVersion); err != nil {
		return SessionToken{}, err
	}
	if strings.TrimSpace(clientNonce) == "" {
		return SessionToken{}, RefusalError{
			Reason: RefusalReasonUnauthorizedCapability,
			Detail: "missing handshake nonce",
		}
	}

	tokenValue, err := randomToken()
	if err != nil {
		return SessionToken{}, err
	}
	token := SessionToken{
		Value:     tokenValue,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, seen := h.nonces[clientNonce]; seen {
		return SessionToken{}, RefusalError{
			Reason: RefusalReasonNonceReplay,
			Detail: fmt.Sprintf("nonce %s reused", clientNonce),
		}
	}
	h.nonces[clientNonce] = struct{}{}
	h.tokens[token.Value] = token
	return token, nil
}

func (h *HandshakeMiddleware) ensureSupportedVersion(clientVersion string) error {
	clientValue, err := parseVersionInt(clientVersion)
	if err != nil {
		return RefusalError{
			Reason: RefusalReasonUnauthorizedCapability,
			Detail: fmt.Sprintf("client version %s is not supported", clientVersion),
		}
	}

	minValue, err := parseVersionInt(h.minVersion)
	if err != nil {
		return RefusalError{
			Reason: RefusalReasonUnauthorizedCapability,
			Detail: fmt.Sprintf("minimum version %s is invalid", h.minVersion),
		}
	}

	if clientValue < minValue {
		return RefusalError{
			Reason: RefusalReasonVersionDowngrade,
			Detail: fmt.Sprintf("client version %s below minimum %s", clientVersion, h.minVersion),
		}
	}

	if clientVersion != h.version {
		return RefusalError{
			Reason: RefusalReasonUnauthorizedCapability,
			Detail: fmt.Sprintf("client version %s is not supported", clientVersion),
		}
	}
	return nil
}

func parseVersionInt(version string) (int, error) {
	var digits strings.Builder
	for _, r := range version {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		} else if digits.Len() > 0 {
			break
		}
	}
	if digits.Len() == 0 {
		return 0, fmt.Errorf("invalid version %s", version)
	}
	return strconv.Atoi(digits.String())
}

// AuthenticateSession verifies a previously issued token.
func (h *HandshakeMiddleware) AuthenticateSession(value string) (SessionToken, error) {
	h.mu.Lock()
	token, ok := h.tokens[value]
	h.mu.Unlock()

	if !ok || !token.Valid() {
		return SessionToken{}, RefusalError{
			Reason: RefusalReasonInvalidToken,
			Detail: "token missing or expired",
		}
	}

	return token, nil
}

func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
