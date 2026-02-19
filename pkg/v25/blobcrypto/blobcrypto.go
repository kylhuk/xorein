package blobcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

// Scope describes the key domain used for envelope wrapping.
type Scope string

const (
	ScopeSpaceAsset Scope = "space_asset"
	ScopeDMSession  Scope = "dm_session"
)

type RefusalCode string

const (
	RefusalCodeMissingKeyMaterial RefusalCode = "missing_key_material"
	RefusalCodeInvalidEnvelope    RefusalCode = "invalid_envelope"
	RefusalCodeKeyRevoked         RefusalCode = "key_revoked"
	RefusalCodeAuthFailure        RefusalCode = "auth_failure"
)

// RefusalError is a deterministic refusal surface for blob envelopes.
type RefusalError struct {
	Code   RefusalCode
	Reason string
}

func (e RefusalError) Error() string {
	if e.Code == "" && e.Reason == "" {
		return "refusal"
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Reason)
}

func newRefusalError(code RefusalCode, reason string) *RefusalError {
	return &RefusalError{Code: code, Reason: reason}
}

// Envelope carries the wrapped DEK for a given asset scope.
type Envelope struct {
	Scope      Scope  `json:"scope"`
	KeyID      string `json:"keyId"`
	WrappedDEK []byte `json:"wrappedDek"`
	Nonce      []byte `json:"nonce"`
}

// KeyMaterial identifies wrapping material for a Scope.
type KeyMaterial struct {
	Scope Scope
	KeyID string
	Key   []byte
}

// RotationHook allows callers to mark scoped key identifiers as revoked.
type RotationHook func(scope Scope, keyID string) error

// KeyRegistry stores scope → key relationships and exposes rotation hooks.
type KeyRegistry struct {
	keys     map[Scope]KeyMaterial
	rotation RotationHook
}

// NewKeyRegistry returns a registry with an optional rotation hook.
func NewKeyRegistry(rotation RotationHook) *KeyRegistry {
	return &KeyRegistry{keys: make(map[Scope]KeyMaterial), rotation: rotation}
}

// Register binds key material to a scope.
func (kr *KeyRegistry) Register(material KeyMaterial) {
	if kr == nil || material.Scope == "" {
		return
	}
	if kr.keys == nil {
		kr.keys = make(map[Scope]KeyMaterial)
	}
	kr.keys[material.Scope] = material
}

func (kr *KeyRegistry) fetch(scope Scope) (KeyMaterial, error) {
	if kr == nil || kr.keys == nil {
		return KeyMaterial{}, newRefusalError(RefusalCodeMissingKeyMaterial, "registry not configured")
	}
	material, ok := kr.keys[scope]
	if !ok {
		return KeyMaterial{}, newRefusalError(RefusalCodeMissingKeyMaterial, fmt.Sprintf("no key for scope %s", scope))
	}
	return material, nil
}

func (kr *KeyRegistry) checkRotation(scope Scope, keyID string) error {
	if kr == nil || kr.rotation == nil {
		return nil
	}
	return kr.rotation(scope, keyID)
}

// WrapSpaceDEK wraps the DEK under the space asset key.
func WrapSpaceDEK(registry *KeyRegistry, dek []byte) (*Envelope, error) {
	return wrapDEK(registry, ScopeSpaceAsset, dek)
}

// WrapDMSessionDEK wraps the DEK under the DM session key.
func WrapDMSessionDEK(registry *KeyRegistry, dek []byte) (*Envelope, error) {
	return wrapDEK(registry, ScopeDMSession, dek)
}

func wrapDEK(registry *KeyRegistry, scope Scope, dek []byte) (*Envelope, error) {
	material, err := registry.fetch(scope)
	if err != nil {
		return nil, err
	}
	return encryptDEK(material, dek)
}

func encryptDEK(material KeyMaterial, dek []byte) (*Envelope, error) {
	if len(material.Key) == 0 {
		return nil, fmt.Errorf("key material missing bytes")
	}
	block, err := aes.NewCipher(material.Key)
	if err != nil {
		return nil, fmt.Errorf("invalid key material: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to init AEAD: %w", err)
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := aead.Seal(nil, nonce, dek, []byte(material.Scope))
	return &Envelope{
		Scope:      material.Scope,
		KeyID:      material.KeyID,
		WrappedDEK: ciphertext,
		Nonce:      nonce,
	}, nil
}

// UnwrapDEK validates the envelope and returns the plaintext DEK.
func UnwrapDEK(registry *KeyRegistry, envelope *Envelope) ([]byte, error) {
	if envelope == nil {
		return nil, newRefusalError(RefusalCodeInvalidEnvelope, "missing envelope")
	}
	material, err := registry.fetch(envelope.Scope)
	if err != nil {
		return nil, err
	}
	if material.KeyID != envelope.KeyID {
		return nil, newRefusalError(RefusalCodeInvalidEnvelope, fmt.Sprintf("unexpected key id %s", envelope.KeyID))
	}
	if err := registry.checkRotation(envelope.Scope, envelope.KeyID); err != nil {
		return nil, newRefusalError(RefusalCodeKeyRevoked, err.Error())
	}
	block, err := aes.NewCipher(material.Key)
	if err != nil {
		return nil, newRefusalError(RefusalCodeInvalidEnvelope, "invalid key material")
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, newRefusalError(RefusalCodeInvalidEnvelope, "failed to init aead")
	}
	plaintext, err := aead.Open(nil, envelope.Nonce, envelope.WrappedDEK, []byte(envelope.Scope))
	if err != nil {
		return nil, newRefusalError(RefusalCodeAuthFailure, "envelope authentication failed")
	}
	return plaintext, nil
}
