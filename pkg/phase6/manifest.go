package phase6

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	// ManifestVersionV1 is the initial manifest schema version.
	ManifestVersionV1 = 1
)

var (
	ErrManifestNil               = errors.New("manifest required")
	ErrManifestIdentityRequired  = errors.New("identity required for signing")
	ErrManifestServerIDRequired  = errors.New("server id required")
	ErrManifestVersionInvalid    = errors.New("manifest version must be >= 1")
	ErrManifestUpdatedAtRequired = errors.New("updated_at required")
	ErrManifestStale             = errors.New("manifest is stale")
)

// Capabilities enumerates the surface that a server promises to offer once
// joined and the membership state exposes hooks for chat/voice activation.
type Capabilities struct {
	Chat  bool `json:"chat"`
	Voice bool `json:"voice"`
}

// Manifest captures server metadata that is signed/deterministic across
// runtime instances of the server, allowing remote peers to resolve
// compatibility requirements through the publish/resolve service.
type Manifest struct {
	ServerID     string       `json:"server_id"`
	Identity     string       `json:"identity"`
	Version      int          `json:"version"`
	Description  string       `json:"description"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Capabilities Capabilities `json:"capabilities"`
	Signature    string       `json:"signature"`
}

// ValidateFields enforces mandatory fields and baseline versioning constraints.
func (m *Manifest) ValidateFields() error {
	if m == nil {
		return ErrManifestNil
	}
	if m.ServerID == "" {
		return ErrManifestServerIDRequired
	}
	if m.Version < ManifestVersionV1 {
		return ErrManifestVersionInvalid
	}
	if m.UpdatedAt.IsZero() {
		return ErrManifestUpdatedAtRequired
	}
	return nil
}

// ValidateFreshness rejects stale manifests compared to the provided reference time
// and maximum tolerated age.
func (m *Manifest) ValidateFreshness(reference time.Time, maxAge time.Duration) error {
	if err := m.ValidateFields(); err != nil {
		return err
	}
	if maxAge <= 0 {
		return nil
	}
	if reference.IsZero() {
		reference = time.Now().UTC()
	}
	if reference.Sub(m.UpdatedAt.UTC()) > maxAge {
		return ErrManifestStale
	}
	return nil
}

// Serialize returns a deterministic byte representation of the manifest.
func (m *Manifest) Serialize() ([]byte, error) {
	if err := m.ValidateFields(); err != nil {
		return nil, err
	}

	type canonical struct {
		ServerID     string       `json:"server_id"`
		Identity     string       `json:"identity"`
		Version      int          `json:"version"`
		Description  string       `json:"description"`
		UpdatedAt    string       `json:"updated_at"`
		Capabilities Capabilities `json:"capabilities"`
	}

	payload := canonical{
		ServerID:     m.ServerID,
		Identity:     m.Identity,
		Version:      m.Version,
		Description:  m.Description,
		UpdatedAt:    m.UpdatedAt.UTC().Format(time.RFC3339Nano),
		Capabilities: m.Capabilities,
	}

	return json.Marshal(payload)
}

// Sign deterministically signs the manifest using the provided identity string.
// The identity string is stored on the manifest so that peers can verify it.
func (m *Manifest) Sign(identity string) (string, error) {
	if m == nil {
		return "", ErrManifestNil
	}
	if identity == "" {
		return "", ErrManifestIdentityRequired
	}

	m.Identity = identity
	serialized, err := m.Serialize()
	if err != nil {
		return "", fmt.Errorf("serialize before sign: %w", err)
	}

	digest := sha256.Sum256(append(serialized, []byte(identity)...))
	sig := hex.EncodeToString(digest[:])
	m.Signature = sig
	return sig, nil
}

// ValidateSignature recreates the deterministic signature and compares it.
func (m *Manifest) ValidateSignature(identity string) bool {
	if m == nil || identity == "" {
		return false
	}
	current, err := m.Serialize()
	if err != nil {
		return false
	}
	digest := sha256.Sum256(append(current, []byte(identity)...))
	return hex.EncodeToString(digest[:]) == m.Signature
}

// Clone deep copies the manifest to avoid shared mutation between callers.
func (m *Manifest) Clone() *Manifest {
	if m == nil {
		return nil
	}
	clone := *m
	clone.Capabilities = m.Capabilities
	clone.UpdatedAt = m.UpdatedAt
	return &clone
}

// ServerState captures the local representation of a server that produced a
// signed manifest, allowing CLI tooling to emit state and hook into future
// chat/voice activation flows.
type ServerState struct {
	Manifest      *Manifest
	PublishedAt   time.Time
	LocalMetadata map[string]string
}

// NewServerState returns a new ServerState for the provided manifest.
func NewServerState(m *Manifest) *ServerState {
	return &ServerState{
		Manifest:      m,
		PublishedAt:   time.Now().UTC(),
		LocalMetadata: map[string]string{},
	}
}

// AddMetadata attaches simple key-value state to the server state.
func (s *ServerState) AddMetadata(key, value string) {
	if s == nil {
		return
	}
	if s.LocalMetadata == nil {
		s.LocalMetadata = map[string]string{}
	}
	s.LocalMetadata[key] = value
}
