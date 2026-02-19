package blobcrypto

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
)

const dekSize = 32

// GenerateDEK produces a random per-blob DEK.
func GenerateDEK() ([]byte, error) {
	buf := make([]byte, dekSize)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// DeriveChunkNonce produces a deterministic nonce from blob hash + chunk index.
func DeriveChunkNonce(blobHash []byte, chunkIndex uint64, nonceSize int) ([]byte, error) {
	if len(blobHash) == 0 {
		return nil, fmt.Errorf("blob hash required")
	}
	if nonceSize <= 0 {
		return nil, fmt.Errorf("nonce size must be positive")
	}
	sum := sha256.Sum256(blobHash)
	var idx [8]byte
	binary.BigEndian.PutUint64(idx[:], chunkIndex)
	h := sha256.New()
	h.Write(sum[:])
	h.Write(idx[:])
	digest := h.Sum(nil)
	if nonceSize > len(digest) {
		return nil, fmt.Errorf("nonce size %d exceeds %d", nonceSize, len(digest))
	}
	nonce := make([]byte, nonceSize)
	copy(nonce, digest[:nonceSize])
	return nonce, nil
}

// EncryptChunk seals chunk plaintext using deterministic nonce derivation.
func EncryptChunk(aead cipher.AEAD, blobHash []byte, chunkIndex uint64, plaintext, associatedData []byte) ([]byte, error) {
	nonce, err := DeriveChunkNonce(blobHash, chunkIndex, aead.NonceSize())
	if err != nil {
		return nil, err
	}
	return aead.Seal(nil, nonce, plaintext, associatedData), nil
}

// DecryptChunk opens a chunk ciphertext that was sealed with EncryptChunk.
func DecryptChunk(aead cipher.AEAD, blobHash []byte, chunkIndex uint64, ciphertext, associatedData []byte) ([]byte, error) {
	nonce, err := DeriveChunkNonce(blobHash, chunkIndex, aead.NonceSize())
	if err != nil {
		return nil, err
	}
	return aead.Open(nil, nonce, ciphertext, associatedData)
}
