package blobcrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"testing"
)

func requireRefusalCode(t *testing.T, err error, want RefusalCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected refusal")
	}
	re, ok := err.(*RefusalError)
	if !ok {
		t.Fatalf("unexpected error type %T", err)
	}
	if re.Code != want {
		t.Fatalf("unexpected refusal code %s", re.Code)
	}
}

func newTestRegistry(hook RotationHook) *KeyRegistry {
	registry := NewKeyRegistry(hook)
	registry.Register(KeyMaterial{
		Scope: ScopeSpaceAsset,
		KeyID: "space-key",
		Key:   bytes.Repeat([]byte{1}, 32),
	})
	registry.Register(KeyMaterial{
		Scope: ScopeDMSession,
		KeyID: "dm-key",
		Key:   bytes.Repeat([]byte{2}, 32),
	})
	return registry
}

func TestWrapUnwrapRoundtrip(t *testing.T) {
	registry := newTestRegistry(nil)
	dek := bytes.Repeat([]byte{3}, dekSize)
	envelope, err := WrapSpaceDEK(registry, dek)
	if err != nil {
		t.Fatalf("wrap failed: %v", err)
	}
	decoded, err := UnwrapDEK(registry, envelope)
	if err != nil {
		t.Fatalf("unwrap failed: %v", err)
	}
	if !bytes.Equal(dek, decoded) {
		t.Fatalf("dek mismatch")
	}
}

func TestWrapSpaceMissingKey(t *testing.T) {
	registry := NewKeyRegistry(nil)
	_, err := WrapSpaceDEK(registry, bytes.Repeat([]byte{4}, dekSize))
	requireRefusalCode(t, err, RefusalCodeMissingKeyMaterial)
}

func TestUnwrapAuthFailure(t *testing.T) {
	registry := newTestRegistry(nil)
	dek := bytes.Repeat([]byte{5}, dekSize)
	envelope, err := WrapSpaceDEK(registry, dek)
	if err != nil {
		t.Fatalf("wrap failed: %v", err)
	}
	envelope.WrappedDEK[0] ^= 0xAA
	_, err = UnwrapDEK(registry, envelope)
	requireRefusalCode(t, err, RefusalCodeAuthFailure)
}

func TestUnwrapKeyRevoked(t *testing.T) {
	registry := newTestRegistry(func(scope Scope, keyID string) error {
		if keyID == "space-key" {
			return errors.New("revoked")
		}
		return nil
	})
	dek := bytes.Repeat([]byte{6}, dekSize)
	envelope, err := WrapSpaceDEK(registry, dek)
	if err != nil {
		t.Fatalf("wrap failed: %v", err)
	}
	_, err = UnwrapDEK(registry, envelope)
	requireRefusalCode(t, err, RefusalCodeKeyRevoked)
}

func TestDeterministicChunkNonce(t *testing.T) {
	blobHash := []byte("blob-hash")
	nonce1, err := DeriveChunkNonce(blobHash, 2, 12)
	if err != nil {
		t.Fatalf("derive failed: %v", err)
	}
	nonce2, err := DeriveChunkNonce(blobHash, 2, 12)
	if err != nil {
		t.Fatalf("derive failed: %v", err)
	}
	if !bytes.Equal(nonce1, nonce2) {
		t.Fatalf("nonce mismatch")
	}
	block, err := aes.NewCipher(bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatalf("cipher init: %v", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("gcm init: %v", err)
	}
	plaintext := []byte("chunk payload")
	ciphertext, err := EncryptChunk(aead, blobHash, 2, plaintext, nil)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	recovered, err := DecryptChunk(aead, blobHash, 2, ciphertext, nil)
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}
	if !bytes.Equal(plaintext, recovered) {
		t.Fatalf("chunk roundtrip")
	}
}
