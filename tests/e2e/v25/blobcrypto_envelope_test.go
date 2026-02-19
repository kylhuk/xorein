package v25

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"testing"

	"github.com/aether/code_aether/pkg/v25/blobcrypto"
)

func TestBlobCryptoEnvelopeRoundtrip(t *testing.T) {
	registry := blobcrypto.NewKeyRegistry(nil)
	registry.Register(blobcrypto.KeyMaterial{Scope: blobcrypto.ScopeSpaceAsset, KeyID: "space-key", Key: bytes.Repeat([]byte{0x10}, 32)})
	registry.Register(blobcrypto.KeyMaterial{Scope: blobcrypto.ScopeDMSession, KeyID: "dm-key", Key: bytes.Repeat([]byte{0x20}, 32)})

	dek, err := blobcrypto.GenerateDEK()
	if err != nil {
		t.Fatalf("dek generation failed: %v", err)
	}

	envelope, err := blobcrypto.WrapSpaceDEK(registry, dek)
	if err != nil {
		t.Fatalf("envelope wrap failed: %v", err)
	}

	plaintextDek, err := blobcrypto.UnwrapDEK(registry, envelope)
	if err != nil {
		t.Fatalf("envelope unwrap failed: %v", err)
	}
	if !bytes.Equal(dek, plaintextDek) {
		t.Fatalf("dek mismatch")
	}

	block, err := aes.NewCipher(bytes.Repeat([]byte{0x30}, 32))
	if err != nil {
		t.Fatalf("cipher init: %v", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("gcm init: %v", err)
	}
	payload := []byte("chunk payload")
	ciphertext, err := blobcrypto.EncryptChunk(aead, []byte("blob-id"), 1, payload, nil)
	if err != nil {
		t.Fatalf("chunk encrypt: %v", err)
	}
	recovered, err := blobcrypto.DecryptChunk(aead, []byte("blob-id"), 1, ciphertext, nil)
	if err != nil {
		t.Fatalf("chunk decrypt: %v", err)
	}
	if !bytes.Equal(payload, recovered) {
		t.Fatalf("chunk mismatch")
	}
}
